package services

import (
	"fmt"
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

const maxConsultasSimultaneas = 5

func queryConTipoFecha(query, valor string) (string, error) {
	columna, err := ColumnaTipoFecha(valor, string(FechaEmision))
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(query, "{{FECHA_COLUMN}}", columna), nil
}

func tipoFechaOpcional(valores []string) string {
	if len(valores) == 0 {
		return ""
	}
	return valores[0]
}

type ResumenContableService struct {
	ConsultasRepo repositories.ConsutasRepository
}

func NewResumenContableService(r *repositories.ConsutasRepository) *ResumenContableService {
	return &ResumenContableService{ConsultasRepo: *r}
}

// Sin NOLOCK: el débito fiscal requiere lectura consistente.
const libroVentasIvaQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    CAST(sdf.fecha_emision AS date)                                     AS fecha,
    ss.nombre                                                           AS sucursal,
    sdf.numero_factura,
    sdf.cuf,
    sdf.codigo_recepcion_sin,
    sdf.numero_documento,
    sdf.complemento,
    sdf.nombre_razon_social,
    sdf.monto_total,
    ISNULL(sdf.monto_ice_especifico,0) + ISNULL(sdf.monto_ice_porcentual,0) AS ice,
    ISNULL(sdf.monto_descuento,0) + ISNULL(sdf.descuento_adicional,0)      AS descuentos,
    sdf.monto_total_sujeto_iva                                          AS base_debito_fiscal,
    ROUND(sdf.monto_total_sujeto_iva * 0.13, 2)                         AS debito_fiscal,
    CASE WHEN sdf.estado_documento_fiscal = 'ANULADO' THEN 'A' ELSE 'V' END AS estado_libro,
    sdf.estado_documento_fiscal
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
JOIN FacturacionNaabol.dbo.sfe_sucursal ss ON ss.id = sdf.id_sfe_sucursal
WHERE {{FECHA_COLUMN}} >= @FechaDesde
  AND {{FECHA_COLUMN}} <  @FechaHasta
  AND sdf.estado_documento_fiscal IN ('VERIFICADO', 'ANULADO')
ORDER BY ss.nombre, sdf.numero_factura;
`

const libroVentasIvaResumenQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    ss.nombre AS sucursal,
    SUM(CASE WHEN sdf.estado_documento_fiscal = 'VERIFICADO' THEN 1 ELSE 0 END) AS facturas_validas,
    SUM(CASE WHEN sdf.estado_documento_fiscal = 'ANULADO'   THEN 1 ELSE 0 END) AS facturas_anuladas,
    SUM(CASE WHEN sdf.estado_documento_fiscal = 'VERIFICADO' THEN sdf.monto_total ELSE 0 END) AS monto_total,
    SUM(CASE WHEN sdf.estado_documento_fiscal = 'VERIFICADO' THEN sdf.monto_total_sujeto_iva ELSE 0 END) AS base_debito_fiscal,
    ROUND(SUM(CASE WHEN sdf.estado_documento_fiscal = 'VERIFICADO'
                   THEN sdf.monto_total_sujeto_iva ELSE 0 END) * 0.13, 2) AS debito_fiscal
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
JOIN FacturacionNaabol.dbo.sfe_sucursal ss ON ss.id = sdf.id_sfe_sucursal
WHERE {{FECHA_COLUMN}} >= @FechaDesde
  AND {{FECHA_COLUMN}} <  @FechaHasta
  AND sdf.estado_documento_fiscal IN ('VERIFICADO', 'ANULADO')
GROUP BY ss.nombre
ORDER BY ss.nombre;
`

func (s *ResumenContableService) conectarServidor(idServer int64) (*gorm.DB, error) {
	server, err := s.ConsultasRepo.GetServidorById(idServer)
	if err != nil {
		return nil, err
	}
	return abrirServidor(server)
}

func abrirServidor(server *models.DbConnection) (*gorm.DB, error) {
	db, err := gorm.Open(sqlserver.Open(server.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar: %w", err)
	}
	return db, nil
}

func (s *ResumenContableService) LibroVentasIva(idServer int64, fechaDesde, fechaHasta time.Time, tipoFecha ...string) ([]models.LibroVentaIva, error) {
	db, err := s.conectarServidor(idServer)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.LibroVentaIva
	query, err := queryConTipoFecha(libroVentasIvaQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	if err := db.Raw(query, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar el libro de ventas IVA: %w", err)
	}
	return filas, nil
}

func (s *ResumenContableService) LibroVentasIvaResumen(idServer int64, fechaDesde, fechaHasta time.Time, tipoFecha ...string) ([]models.LibroVentaIvaResumenSucursal, error) {
	db, err := s.conectarServidor(idServer)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.LibroVentaIvaResumenSucursal
	query, err := queryConTipoFecha(libroVentasIvaResumenQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	if err := db.Raw(query, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar el resumen del libro de ventas IVA: %w", err)
	}
	return filas, nil
}

// Sin filtro de estado: incluye meses que solo tienen facturas anuladas (válidas en 0).
const resumenMensualDebitoFiscalQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    YEAR({{FECHA_COLUMN}})  AS anio,
    MONTH({{FECHA_COLUMN}}) AS mes,
    SUM(CASE WHEN sdf.estado_documento_fiscal = 'VERIFICADO' THEN 1 ELSE 0 END) AS facturas_validas,
    SUM(CASE WHEN sdf.estado_documento_fiscal = 'ANULADO'   THEN 1 ELSE 0 END) AS facturas_anuladas,
    SUM(CASE WHEN sdf.estado_documento_fiscal = 'VERIFICADO' THEN sdf.monto_total ELSE 0 END) AS total_facturado,
    SUM(CASE WHEN sdf.estado_documento_fiscal = 'VERIFICADO' THEN sdf.monto_total_sujeto_iva ELSE 0 END) AS base_debito_fiscal,
    ROUND(SUM(CASE WHEN sdf.estado_documento_fiscal = 'VERIFICADO'
                   THEN sdf.monto_total_sujeto_iva ELSE 0 END) * 0.13, 2) AS debito_fiscal
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
WHERE {{FECHA_COLUMN}} >= @FechaDesde
  AND {{FECHA_COLUMN}} <  @FechaHasta
GROUP BY YEAR({{FECHA_COLUMN}}), MONTH({{FECHA_COLUMN}})
ORDER BY anio, mes;
`

func (s *ResumenContableService) ResumenMensualDebitoFiscal(idServer int64, fechaDesde, fechaHasta time.Time, tipoFecha ...string) ([]models.ResumenMensualDebitoFiscal, error) {
	db, err := s.conectarServidor(idServer)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.ResumenMensualDebitoFiscal
	query, err := queryConTipoFecha(resumenMensualDebitoFiscalQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	if err := db.Raw(query, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar el resumen mensual de débito fiscal: %w", err)
	}
	return filas, nil
}

// EXISTS evita contar doble una factura con reintentos de anulación; sucursales sin anulaciones salen en 0.
const facturasAnuladasQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    ss.nombre                  AS sucursal,
    COUNT(d.id)                AS facturas,
    ISNULL(SUM(d.monto_total), 0) AS monto
FROM FacturacionNaabol.dbo.sfe_sucursal ss
LEFT JOIN (
    SELECT sdf.id, sdf.id_sfe_sucursal, sdf.monto_total
    FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
    WHERE {{FECHA_COLUMN}} >= @FechaDesde
      AND {{FECHA_COLUMN}} <  @FechaHasta
      AND sdf.estado_documento_fiscal = 'ANULADO'
      AND EXISTS (
          SELECT 1
          FROM FacturacionNaabol.dbo.sfe_anulacion sa
          WHERE sa.id_sfe_documento_fiscal = sdf.id
            AND sa.estado_anulacion = 'VERIFICADO'
      )
) d ON d.id_sfe_sucursal = ss.id
GROUP BY ss.id, ss.nombre
ORDER BY ss.nombre;
`

const facturasValidasQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    ss.nombre                     AS sucursal,
    COUNT(sdf.id)                 AS facturas,
    ISNULL(SUM(sdf.monto_total), 0) AS monto
FROM FacturacionNaabol.dbo.sfe_sucursal ss
LEFT JOIN FacturacionNaabol.dbo.sfe_documento_fiscal sdf
       ON sdf.id_sfe_sucursal = ss.id
      AND {{FECHA_COLUMN}} >= @FechaDesde
      AND {{FECHA_COLUMN}} <  @FechaHasta
      AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY ss.id, ss.nombre
ORDER BY ss.nombre;
`

func (s *ResumenContableService) FacturasAnuladas(ambiente string, fechaDesde, fechaHasta time.Time, tipoFecha ...string) ([]models.ServidorTotales, error) {
	query, err := queryConTipoFecha(facturasAnuladasQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	return s.totalesTodosServidores(ambiente, query, fechaDesde, fechaHasta)
}

func (s *ResumenContableService) FacturasValidas(ambiente string, fechaDesde, fechaHasta time.Time, tipoFecha ...string) ([]models.ServidorTotales, error) {
	query, err := queryConTipoFecha(facturasValidasQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	return s.totalesTodosServidores(ambiente, query, fechaDesde, fechaHasta)
}

// Un servidor caído no tumba el reporte; se informa en su fila.
func (s *ResumenContableService) totalesTodosServidores(ambiente, query string, fechaDesde, fechaHasta time.Time) ([]models.ServidorTotales, error) {
	servidores, err := s.ConsultasRepo.GetServidoresFacturador(ambiente)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo servidores: %w", err)
	}

	resultados := make([]models.ServidorTotales, len(servidores))
	limite := make(chan struct{}, maxConsultasSimultaneas)
	var wg sync.WaitGroup
	for i := range servidores {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			limite <- struct{}{}
			defer func() { <-limite }()
			resultados[i].Servidor = servidores[i].ServerName
			var err error
			resultados[i].Sucursales, err = consultarTotalesSucursal(&servidores[i], query, fechaDesde, fechaHasta)
			if err != nil {
				resultados[i].Error = err.Error()
			}
		}(i)
	}
	wg.Wait()
	return resultados, nil
}

func consultarTotalesSucursal(server *models.DbConnection, query string, fechaDesde, fechaHasta time.Time) ([]models.SucursalTotal, error) {
	db, err := abrirServidor(server)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.SucursalTotal
	if err := db.Raw(query, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al consultar totales por sucursal: %w", err)
	}
	return filas, nil
}

const notasCreditoDebitoQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    CAST(sdf.fecha_emision AS date) AS fecha,
    sdf.numero_factura              AS nro_nota,
    sdf.tipo_documento_sector,
    sdf.numero_factura_original,
    sdf.numero_autorizacion_cuf     AS cuf_factura_original,
    CAST(sdf.fecha_emision_factura AS date) AS fecha_factura_original,
    sdf.numero_documento,
    sdf.nombre_razon_social,
    sdf.monto_total_original,
    sdf.monto_total_devuelto,
    sdf.monto_total_conciliado,
    sdf.debito_fiscal_iva,
    sdf.credito_fiscal_iva,
    sdf.estado_documento_fiscal
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
WHERE {{FECHA_COLUMN}} >= @FechaDesde
  AND {{FECHA_COLUMN}} <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
  AND (sdf.tipo_documento_sector IN (24, 29)
       OR sdf.numero_factura_original IS NOT NULL
       OR sdf.numero_autorizacion_cuf IS NOT NULL)
ORDER BY {{FECHA_COLUMN}}, sdf.numero_factura;
`

func (s *ResumenContableService) NotasCreditoDebito(idServer int64, fechaDesde, fechaHasta time.Time, tipoFecha ...string) ([]models.NotaCreditoDebito, error) {
	db, err := s.conectarServidor(idServer)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.NotaCreditoDebito
	query, err := queryConTipoFecha(notasCreditoDebitoQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	if err := db.Raw(query, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar notas de crédito/débito y conciliación: %w", err)
	}
	return filas, nil
}

const facturacionPorActividadQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    sdf.codigo_actividad_economica,
    sdf.actividad_economica,
    COUNT(*)                        AS facturas,
    SUM(sdf.monto_total)            AS total_facturado,
    SUM(sdf.monto_total_sujeto_iva) AS base_debito_fiscal
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
WHERE {{FECHA_COLUMN}} >= @FechaDesde
  AND {{FECHA_COLUMN}} <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY sdf.codigo_actividad_economica, sdf.actividad_economica
ORDER BY total_facturado DESC;
`

func (s *ResumenContableService) FacturacionPorActividad(idServer int64, fechaDesde, fechaHasta time.Time, tipoFecha ...string) ([]models.FacturacionPorActividad, error) {
	db, err := s.conectarServidor(idServer)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.FacturacionPorActividad
	query, err := queryConTipoFecha(facturacionPorActividadQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	if err := db.Raw(query, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar facturación por actividad económica: %w", err)
	}
	return filas, nil
}

const ingresosPorProductoQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    sddf.codigo_producto_sfe,
    sddf.codigo_producto_sin,
    MAX(sddf.descripcion)                       AS descripcion_factura,
    MAX(ps.descripcion_producto_sin)            AS descripcion_sin,
    SUM(sddf.cantidad)                          AS cantidad,
    SUM(sddf.sub_total)                         AS subtotal,
    SUM(ISNULL(sddf.monto_descuento_detalle,0)) AS descuento_item
FROM FacturacionNaabol.dbo.sfe_detalle_documento_fiscal sddf
JOIN FacturacionNaabol.dbo.sfe_documento_fiscal sdf ON sdf.id = sddf.id_sfe_documento_fiscal
JOIN FacturacionNaabol.dbo.sfe_sucursal ss           ON ss.id = sdf.id_sfe_sucursal
OUTER APPLY (
    SELECT TOP 1 p.descripcion_producto_sin
    FROM   FacturacionNaabol.dbo.sfe_parametrica_producto_sin p
    WHERE  p.codigo_producto_sfe = sddf.codigo_producto_sfe
      AND  p.id_sfe_empresa      = ss.id_sfe_empresa
) ps
WHERE {{FECHA_COLUMN}} >= @FechaDesde
  AND {{FECHA_COLUMN}} <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY sddf.codigo_producto_sfe, sddf.codigo_producto_sin
ORDER BY subtotal DESC;
`

func (s *ResumenContableService) IngresosPorProducto(idServer int64, fechaDesde, fechaHasta time.Time, tipoFecha ...string) ([]models.IngresosPorProducto, error) {
	db, err := s.conectarServidor(idServer)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.IngresosPorProducto
	query, err := queryConTipoFecha(ingresosPorProductoQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	if err := db.Raw(query, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar ingresos por producto: %w", err)
	}
	return filas, nil
}

const ingresosPorSucursalPosQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    ss.codigo_sucursal,
    ss.nombre                       AS sucursal,
    ss.municipio_departamento,
    spv.codigo_pos,
    spv.nombre                      AS punto_venta,
    COUNT(*)                        AS facturas,
    SUM(sdf.monto_total)            AS total_facturado,
    SUM(sdf.monto_total_sujeto_iva) AS base_debito_fiscal
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
JOIN FacturacionNaabol.dbo.sfe_sucursal ss           ON ss.id  = sdf.id_sfe_sucursal
LEFT JOIN FacturacionNaabol.dbo.sfe_punto_venta spv  ON spv.id = sdf.id_sfe_punto_venta
WHERE {{FECHA_COLUMN}} >= @FechaDesde
  AND {{FECHA_COLUMN}} <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY ss.codigo_sucursal, ss.nombre, ss.municipio_departamento,
         spv.codigo_pos, spv.nombre
ORDER BY ss.nombre, spv.codigo_pos;
`

func (s *ResumenContableService) IngresosPorSucursalPos(idServer int64, fechaDesde, fechaHasta time.Time, tipoFecha ...string) ([]models.IngresosPorSucursalPos, error) {
	db, err := s.conectarServidor(idServer)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.IngresosPorSucursalPos
	query, err := queryConTipoFecha(ingresosPorSucursalPosQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	if err := db.Raw(query, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar ingresos por sucursal/punto de venta: %w", err)
	}
	return filas, nil
}

const ingresosPorMetodoPagoQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    sdf.metodo_pago          AS codigo_metodo_pago,
    MAX(mp.descripcion_sin)  AS metodo_pago,
    COUNT(*)                 AS facturas,
    SUM(sdf.monto_total)     AS total_facturado
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
JOIN FacturacionNaabol.dbo.sfe_sucursal ss ON ss.id = sdf.id_sfe_sucursal
OUTER APPLY (
    SELECT TOP 1 p.descripcion_sin
    FROM   FacturacionNaabol.dbo.sfe_parametrica_sin p
    WHERE  p.codigo_clasificador_sin = sdf.metodo_pago
      AND  p.id_sfe_empresa          = ss.id_sfe_empresa
      AND  p.tipo_parametrica LIKE '%PAGO%'
) mp
WHERE {{FECHA_COLUMN}} >= @FechaDesde
  AND {{FECHA_COLUMN}} <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY sdf.metodo_pago
ORDER BY total_facturado DESC;
`

func (s *ResumenContableService) IngresosPorMetodoPago(idServer int64, fechaDesde, fechaHasta time.Time, tipoFecha ...string) ([]models.IngresosPorMetodoPago, error) {
	db, err := s.conectarServidor(idServer)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.IngresosPorMetodoPago
	query, err := queryConTipoFecha(ingresosPorMetodoPagoQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	if err := db.Raw(query, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar ingresos por método de pago: %w", err)
	}
	return filas, nil
}

const ingresosPorMonedaQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    sdf.codigo_moneda,
    MAX(mo.descripcion_sin)                      AS moneda,
    sdf.tipo_cambio,
    sdf.tipo_cambio_oficial,
    COUNT(*)                                     AS facturas,
    SUM(sdf.monto_total_moneda)                  AS total_en_moneda_original,
    SUM(sdf.monto_total)                         AS total_bs,
    SUM(ISNULL(sdf.ingreso_diferencia_cambio,0)) AS diferencia_cambio_abs
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
JOIN FacturacionNaabol.dbo.sfe_sucursal ss ON ss.id = sdf.id_sfe_sucursal
OUTER APPLY (
    SELECT TOP 1 p.descripcion_sin
    FROM   FacturacionNaabol.dbo.sfe_parametrica_sin p
    WHERE  p.codigo_clasificador_sin = sdf.codigo_moneda
      AND  p.id_sfe_empresa          = ss.id_sfe_empresa
      AND  p.tipo_parametrica LIKE '%MONEDA%'
) mo
WHERE {{FECHA_COLUMN}} >= @FechaDesde
  AND {{FECHA_COLUMN}} <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY sdf.codigo_moneda, sdf.tipo_cambio, sdf.tipo_cambio_oficial
ORDER BY total_bs DESC;
`

func (s *ResumenContableService) IngresosPorMoneda(idServer int64, fechaDesde, fechaHasta time.Time, tipoFecha ...string) ([]models.IngresosPorMoneda, error) {
	db, err := s.conectarServidor(idServer)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.IngresosPorMoneda
	query, err := queryConTipoFecha(ingresosPorMonedaQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	if err := db.Raw(query, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar ingresos por moneda: %w", err)
	}
	return filas, nil
}

// TOP (@Top) acota el reporte: puede haber muchos clientes distintos.
const facturacionPorClienteQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?
declare @Top        int       = ?

SELECT TOP (@Top)
    sdf.numero_documento,
    sdf.nombre_razon_social,
    COUNT(*)                           AS facturas,
    SUM(sdf.monto_total)               AS total_facturado,
    MIN(CAST(sdf.fecha_emision AS date)) AS primera_factura,
    MAX(CAST(sdf.fecha_emision AS date)) AS ultima_factura
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
WHERE {{FECHA_COLUMN}} >= @FechaDesde
  AND {{FECHA_COLUMN}} <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY sdf.numero_documento, sdf.nombre_razon_social
ORDER BY total_facturado DESC;
`

func (s *ResumenContableService) FacturacionPorCliente(idServer int64, fechaDesde, fechaHasta time.Time, top int, tipoFecha ...string) ([]models.FacturacionPorCliente, error) {
	db, err := s.conectarServidor(idServer)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.FacturacionPorCliente
	query, err := queryConTipoFecha(facturacionPorClienteQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	if err := db.Raw(query, fechaDesde, fechaHasta, top).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar facturación por cliente: %w", err)
	}
	return filas, nil
}

const emisionPorUsuarioQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    sdf.usuario_emision,
    ss.nombre                       AS sucursal,
    CAST({{FECHA_COLUMN}} AS date) AS dia,
    COUNT(*)                        AS facturas,
    SUM(sdf.monto_total)            AS total_facturado
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
JOIN FacturacionNaabol.dbo.sfe_sucursal ss ON ss.id = sdf.id_sfe_sucursal
WHERE {{FECHA_COLUMN}} >= @FechaDesde
  AND {{FECHA_COLUMN}} <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY sdf.usuario_emision, ss.nombre, CAST({{FECHA_COLUMN}} AS date)
ORDER BY dia, sdf.usuario_emision;
`

func (s *ResumenContableService) EmisionPorUsuario(idServer int64, fechaDesde, fechaHasta time.Time, tipoFecha ...string) ([]models.EmisionPorUsuario, error) {
	db, err := s.conectarServidor(idServer)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.EmisionPorUsuario
	query, err := queryConTipoFecha(emisionPorUsuarioQuery, tipoFechaOpcional(tipoFecha))
	if err != nil {
		return nil, err
	}
	if err := db.Raw(query, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar emisión por usuario: %w", err)
	}
	return filas, nil
}
