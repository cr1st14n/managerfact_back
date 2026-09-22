package services

import (
	"fmt"
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// ResumenContableService agrupa los reportes contables/tributarios de
// ClicReportes.md llevados a la app (Libro Ventas IVA es el primero; los
// siguientes de ese documento se van agregando acá con su propio método).
type ResumenContableService struct {
	ConsultasRepo repositories.ConsutasRepository
}

func NewResumenContableService(r *repositories.ConsutasRepository) *ResumenContableService {
	return &ResumenContableService{ConsultasRepo: *r}
}

// libroVentasIvaQuery es el Libro de Ventas IVA: todas las facturas
// (VERIFICADO y ANULADO) de TODAS las sucursales de la empresa en el
// período, con el débito fiscal ya calculado. La base es
// monto_total_sujeto_iva, no monto_total: el total puede incluir conceptos
// fuera de la base (ICE, gift card, otros pagos no sujetos). Ver
// ClicReportes.md sección 1.
//
// A diferencia del resto de consultas de este archivo (y de
// consultas_service.go), NO usa WITH (NOLOCK) a propósito: es una cifra
// contable y para el débito fiscal se prefiere lectura consistente a la
// última fila que haya pasado por la conexión.
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
WHERE sdf.fecha_emision >= @FechaDesde
  AND sdf.fecha_emision <  @FechaHasta
  AND sdf.estado_documento_fiscal IN ('VERIFICADO', 'ANULADO')
ORDER BY ss.nombre, sdf.numero_factura;
`

// libroVentasIvaResumenQuery es el mismo débito fiscal que
// libroVentasIvaQuery pero agregado por sucursal (GROUP BY) en vez de
// factura por factura: es lo que se muestra en pantalla. El detalle
// completo se pide aparte y solo para exportarlo a Excel (ver
// LibroVentasIva), nunca para pintarlo en una tabla -- con miles de
// facturas en el período, esa tabla sería impracticable en el navegador.
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
WHERE sdf.fecha_emision >= @FechaDesde
  AND sdf.fecha_emision <  @FechaHasta
  AND sdf.estado_documento_fiscal IN ('VERIFICADO', 'ANULADO')
GROUP BY ss.nombre
ORDER BY ss.nombre;
`

// conectarServidor abre la conexión SQL Server de la conexión idServer. El
// caller es responsable de cerrar sqlDB (defer sqlDB.Close()).
func (s *ResumenContableService) conectarServidor(idServer int64) (*gorm.DB, error) {
	server, err := s.ConsultasRepo.GetServidorById(idServer)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlserver.Open(server.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar: %w", err)
	}
	return db, nil
}

// LibroVentasIva conecta al servidor de la empresa (idServer) y devuelve el
// detalle factura por factura para [fechaDesde, fechaHasta) -- hasta
// exclusivo, igual que el resto de reportes por fecha de este proyecto.
// Pensado solo para la descarga a Excel: para mostrar en pantalla usar
// LibroVentasIvaResumen.
func (s *ResumenContableService) LibroVentasIva(idServer int64, fechaDesde, fechaHasta time.Time) ([]models.LibroVentaIva, error) {
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
	if err := db.Raw(libroVentasIvaQuery, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar el libro de ventas IVA: %w", err)
	}
	return filas, nil
}

// LibroVentasIvaResumen es el total de control por sucursal para
// [fechaDesde, fechaHasta) -- lo que se muestra en pantalla (ver comentario
// de libroVentasIvaResumenQuery).
func (s *ResumenContableService) LibroVentasIvaResumen(idServer int64, fechaDesde, fechaHasta time.Time) ([]models.LibroVentaIvaResumenSucursal, error) {
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
	if err := db.Raw(libroVentasIvaResumenQuery, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar el resumen del libro de ventas IVA: %w", err)
	}
	return filas, nil
}

// resumenMensualDebitoFiscalQuery es el "total de control" del libro: debe
// coincidir con la suma de libroVentasIvaQuery y con lo declarado al SIN. A
// diferencia de libroVentasIvaResumenQuery (que agrupa por sucursal),
// agrupa por año/mes de TODA la empresa -- pensado para revisar varios
// meses de un vistazo, no un período puntual. Sin filtro de estado en el
// WHERE a propósito (igual que ClicReportes.md sección 2): así un mes sin
// ninguna factura VERIFICADA/ANULADA igual aparece con ceros, en vez de
// desaparecer de la lista y esconder el hueco.
const resumenMensualDebitoFiscalQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    YEAR(sdf.fecha_emision)  AS anio,
    MONTH(sdf.fecha_emision) AS mes,
    SUM(CASE WHEN sdf.estado_documento_fiscal = 'VERIFICADO' THEN 1 ELSE 0 END) AS facturas_validas,
    SUM(CASE WHEN sdf.estado_documento_fiscal = 'ANULADO'   THEN 1 ELSE 0 END) AS facturas_anuladas,
    SUM(CASE WHEN sdf.estado_documento_fiscal = 'VERIFICADO' THEN sdf.monto_total ELSE 0 END) AS total_facturado,
    SUM(CASE WHEN sdf.estado_documento_fiscal = 'VERIFICADO' THEN sdf.monto_total_sujeto_iva ELSE 0 END) AS base_debito_fiscal,
    ROUND(SUM(CASE WHEN sdf.estado_documento_fiscal = 'VERIFICADO'
                   THEN sdf.monto_total_sujeto_iva ELSE 0 END) * 0.13, 2) AS debito_fiscal
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
WHERE sdf.fecha_emision >= @FechaDesde
  AND sdf.fecha_emision <  @FechaHasta
GROUP BY YEAR(sdf.fecha_emision), MONTH(sdf.fecha_emision)
ORDER BY anio, mes;
`

// ResumenMensualDebitoFiscal conecta al servidor de la empresa (idServer) y
// devuelve un total por mes para [fechaDesde, fechaHasta). Ya es un
// agregado chico (una fila por mes), así que no necesita una versión
// "detalle solo para Excel" como LibroVentasIva: lo que se muestra en
// pantalla es directamente exportable.
func (s *ResumenContableService) ResumenMensualDebitoFiscal(idServer int64, fechaDesde, fechaHasta time.Time) ([]models.ResumenMensualDebitoFiscal, error) {
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
	if err := db.Raw(resumenMensualDebitoFiscalQuery, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar el resumen mensual de débito fiscal: %w", err)
	}
	return filas, nil
}

// facturasAnuladasQuery es "Facturas Anuladas del Período": solicitudes de
// anulación filtradas por FECHA_ENVIO de la anulación, no por la fecha de
// emisión de la factura original (ver comentario de models.FacturaAnuladaSin
// sobre el desfase). Puede haber varias filas por factura (reintentos) --
// no se deduplica, para ver el historial completo. Ver ClicReportes.md
// sección 3.
const facturasAnuladasQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    sdf.numero_factura,
    sdf.cuf,
    CAST(sdf.fecha_emision AS date) AS fecha_emision,
    CAST(sa.fecha_envio    AS date) AS fecha_anulacion,
    sdf.numero_documento,
    sdf.nombre_razon_social,
    sdf.monto_total,
    sa.codigo_motivo,
    sa.estado_anulacion,
    sa.codigo_respuesta_sin,
    sdf.estado_documento_fiscal
FROM FacturacionNaabol.dbo.sfe_anulacion sa
JOIN FacturacionNaabol.dbo.sfe_documento_fiscal sdf ON sdf.id = sa.id_sfe_documento_fiscal
WHERE sa.fecha_envio >= @FechaDesde
  AND sa.fecha_envio <  @FechaHasta
ORDER BY sa.fecha_envio;
`

// FacturasAnuladas conecta al servidor de la empresa (idServer) y devuelve
// las solicitudes de anulación con fecha de envío en
// [fechaDesde, fechaHasta). A diferencia de LibroVentasIva no hace falta una
// versión resumen: el volumen de anulaciones es normalmente una fracción
// chica del total facturado, así que mostrar el detalle directo no
// sobrecarga el front.
func (s *ResumenContableService) FacturasAnuladas(idServer int64, fechaDesde, fechaHasta time.Time) ([]models.FacturaAnuladaSin, error) {
	db, err := s.conectarServidor(idServer)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var filas []models.FacturaAnuladaSin
	if err := db.Raw(facturasAnuladasQuery, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar facturas anuladas del período: %w", err)
	}
	return filas, nil
}

// notasCreditoDebitoQuery es "Notas de Crédito/Débito y Conciliación":
// documentos de sector 24/29 (según catálogo SIN, ver ASUNCION A VALIDAR en
// models.NotaCreditoDebito) o cualquier documento que referencie una
// factura original, como respaldo por si el sector difiere de la
// asunción. DEBITO/CREDITO_FISCAL_IVA ya vienen calculados en el
// documento: no se recalcula el 13% acá. Ver ClicReportes.md sección 4.
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
WHERE sdf.fecha_emision >= @FechaDesde
  AND sdf.fecha_emision <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
  AND (sdf.tipo_documento_sector IN (24, 29)
       OR sdf.numero_factura_original IS NOT NULL
       OR sdf.numero_autorizacion_cuf IS NOT NULL)
ORDER BY sdf.fecha_emision, sdf.numero_factura;
`

// NotasCreditoDebito conecta al servidor de la empresa (idServer) y
// devuelve las notas de crédito/débito y conciliación con fecha de emisión
// en [fechaDesde, fechaHasta). Igual que FacturasAnuladas, es un
// subconjunto chico de la facturación total: se muestra el detalle directo,
// sin una versión resumen aparte.
func (s *ResumenContableService) NotasCreditoDebito(idServer int64, fechaDesde, fechaHasta time.Time) ([]models.NotaCreditoDebito, error) {
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
	if err := db.Raw(notasCreditoDebitoQuery, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar notas de crédito/débito y conciliación: %w", err)
	}
	return filas, nil
}

// facturacionPorActividadQuery es "Facturación por Actividad Económica":
// agrupa por la actividad de CABECERA de la factura (una por documento) --
// ver ASUNCION en models.FacturacionPorActividad sobre por qué no se va por
// detalle. Ya es un agregado chico (una fila por actividad). Ver
// ClicReportes.md sección 5.
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
WHERE sdf.fecha_emision >= @FechaDesde
  AND sdf.fecha_emision <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY sdf.codigo_actividad_economica, sdf.actividad_economica
ORDER BY total_facturado DESC;
`

// FacturacionPorActividad conecta al servidor de la empresa (idServer) y
// devuelve el total por actividad económica para [fechaDesde, fechaHasta).
// Ya es un agregado chico: no necesita una versión "detalle solo para
// Excel" como LibroVentasIva.
func (s *ResumenContableService) FacturacionPorActividad(idServer int64, fechaDesde, fechaHasta time.Time) ([]models.FacturacionPorActividad, error) {
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
	if err := db.Raw(facturacionPorActividadQuery, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar facturación por actividad económica: %w", err)
	}
	return filas, nil
}

// ingresosPorProductoQuery es "Ingresos por Servicio/Producto": el reporte
// que mapea a cuentas de ingreso, agrupado por producto (no por factura,
// ver ASUNCION en models.IngresosPorProducto sobre el descuento adicional
// global). OUTER APPLY TOP 1 para la descripción SIN:
// sfe_parametrica_producto_sin puede tener varias filas por producto y un
// JOIN normal duplicaría cantidades y montos. Ver ClicReportes.md sección 6.
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
WHERE sdf.fecha_emision >= @FechaDesde
  AND sdf.fecha_emision <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY sddf.codigo_producto_sfe, sddf.codigo_producto_sin
ORDER BY subtotal DESC;
`

// IngresosPorProducto conecta al servidor de la empresa (idServer) y
// devuelve el total por producto para [fechaDesde, fechaHasta). Ya es un
// agregado chico: no necesita una versión "detalle solo para Excel".
func (s *ResumenContableService) IngresosPorProducto(idServer int64, fechaDesde, fechaHasta time.Time) ([]models.IngresosPorProducto, error) {
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
	if err := db.Raw(ingresosPorProductoQuery, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar ingresos por producto: %w", err)
	}
	return filas, nil
}

// ingresosPorSucursalPosQuery es "Ingresos por Sucursal y Punto de Venta":
// sirve para separar por aeropuerto/caja. LEFT JOIN a punto de venta (no
// INNER): hay facturas sin POS y con INNER se perderían del total. Ver
// ClicReportes.md sección 7.
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
WHERE sdf.fecha_emision >= @FechaDesde
  AND sdf.fecha_emision <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY ss.codigo_sucursal, ss.nombre, ss.municipio_departamento,
         spv.codigo_pos, spv.nombre
ORDER BY ss.nombre, spv.codigo_pos;
`

// IngresosPorSucursalPos conecta al servidor de la empresa (idServer) y
// devuelve el total por sucursal/punto de venta para
// [fechaDesde, fechaHasta). Ya es un agregado chico.
func (s *ResumenContableService) IngresosPorSucursalPos(idServer int64, fechaDesde, fechaHasta time.Time) ([]models.IngresosPorSucursalPos, error) {
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
	if err := db.Raw(ingresosPorSucursalPosQuery, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar ingresos por sucursal/punto de venta: %w", err)
	}
	return filas, nil
}

// ingresosPorMetodoPagoQuery es "Ingresos por Método de Pago". metodo_pago
// guarda el código del catálogo SIN, no el texto: se resuelve contra
// sfe_parametrica_sin filtrando por empresa (vía sucursal) para no mezclar
// catálogos de distintas empresas. El LIKE '%PAGO%' es una suposición del
// documento original (validar contra la sección 0.3 de ClicReportes.md); si
// no resuelve, el reporte igual sale con el código crudo. Ver
// ClicReportes.md sección 8.
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
WHERE sdf.fecha_emision >= @FechaDesde
  AND sdf.fecha_emision <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY sdf.metodo_pago
ORDER BY total_facturado DESC;
`

// IngresosPorMetodoPago conecta al servidor de la empresa (idServer) y
// devuelve el total por método de pago para [fechaDesde, fechaHasta). Ya es
// un agregado chico.
func (s *ResumenContableService) IngresosPorMetodoPago(idServer int64, fechaDesde, fechaHasta time.Time) ([]models.IngresosPorMetodoPago, error) {
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
	if err := db.Raw(ingresosPorMetodoPagoQuery, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar ingresos por método de pago: %w", err)
	}
	return filas, nil
}

// ingresosPorMonedaQuery es "Ingresos por Moneda y Tipo de Cambio". Mismo
// LIKE de suposición que ingresosPorMetodoPagoQuery (validar contra la
// sección 0.3). Ver ClicReportes.md sección 9.
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
WHERE sdf.fecha_emision >= @FechaDesde
  AND sdf.fecha_emision <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY sdf.codigo_moneda, sdf.tipo_cambio, sdf.tipo_cambio_oficial
ORDER BY total_bs DESC;
`

// IngresosPorMoneda conecta al servidor de la empresa (idServer) y devuelve
// el total por moneda para [fechaDesde, fechaHasta). Ya es un agregado
// chico.
func (s *ResumenContableService) IngresosPorMoneda(idServer int64, fechaDesde, fechaHasta time.Time) ([]models.IngresosPorMoneda, error) {
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
	if err := db.Raw(ingresosPorMonedaQuery, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar ingresos por moneda: %w", err)
	}
	return filas, nil
}

// facturacionPorClienteQuery es "Facturación por Cliente": agrupa por
// número de documento + razón social, no por id de cliente (ver ASUNCION en
// models.FacturacionPorCliente). TOP (@Top) es a propósito: el número de
// clientes distintos puede ser grande, así que se trae como mucho esa
// cantidad, ordenada por total facturado. Ver ClicReportes.md sección 10.
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
WHERE sdf.fecha_emision >= @FechaDesde
  AND sdf.fecha_emision <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY sdf.numero_documento, sdf.nombre_razon_social
ORDER BY total_facturado DESC;
`

// FacturacionPorCliente conecta al servidor de la empresa (idServer) y
// devuelve como mucho "top" clientes (por total facturado) para
// [fechaDesde, fechaHasta).
func (s *ResumenContableService) FacturacionPorCliente(idServer int64, fechaDesde, fechaHasta time.Time, top int) ([]models.FacturacionPorCliente, error) {
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
	if err := db.Raw(facturacionPorClienteQuery, fechaDesde, fechaHasta, top).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar facturación por cliente: %w", err)
	}
	return filas, nil
}

// emisionPorUsuarioQuery es "Emisión por Usuario" (arqueo / control
// interno). usuario_emision es texto libre que manda el sistema
// integrador: tomarlo como pista, no como evidencia (ver
// models.EmisionPorUsuario). Ver ClicReportes.md sección 11.
const emisionPorUsuarioQuery = `
declare @FechaDesde datetime2 = ?
declare @FechaHasta datetime2 = ?

SELECT
    sdf.usuario_emision,
    ss.nombre                       AS sucursal,
    CAST(sdf.fecha_emision AS date) AS dia,
    COUNT(*)                        AS facturas,
    SUM(sdf.monto_total)            AS total_facturado
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
JOIN FacturacionNaabol.dbo.sfe_sucursal ss ON ss.id = sdf.id_sfe_sucursal
WHERE sdf.fecha_emision >= @FechaDesde
  AND sdf.fecha_emision <  @FechaHasta
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
GROUP BY sdf.usuario_emision, ss.nombre, CAST(sdf.fecha_emision AS date)
ORDER BY dia, sdf.usuario_emision;
`

// EmisionPorUsuario conecta al servidor de la empresa (idServer) y devuelve
// el total por usuario/sucursal/día para [fechaDesde, fechaHasta). Pensado
// para períodos cortos (día/semana/mes), como el resto de reportes de
// arqueo.
func (s *ResumenContableService) EmisionPorUsuario(idServer int64, fechaDesde, fechaHasta time.Time) ([]models.EmisionPorUsuario, error) {
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
	if err := db.Raw(emisionPorUsuarioQuery, fechaDesde, fechaHasta).Scan(&filas).Error; err != nil {
		return nil, fmt.Errorf("error al generar emisión por usuario: %w", err)
	}
	return filas, nil
}
