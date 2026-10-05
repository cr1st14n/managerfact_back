package services

import (
	"fmt"
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func toNullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

type ConsultasService struct {
	ConsultasRepo   repositories.ConsutasRepository
	UsuariosRepo    repositories.UsuarioRepository
	SucursalesCache *repositories.ConexionSucursalRepo
}

func NewConsultasService(r *repositories.ConsutasRepository, u *repositories.UsuarioRepository, sc *repositories.ConexionSucursalRepo) *ConsultasService {
	return &ConsultasService{
		ConsultasRepo:   *r,
		UsuariosRepo:    *u,
		SucursalesCache: sc,
	}
}

const dataFacturasQuery = `
declare @NumeroFactura         numeric(19,2) = ?
declare @CodigoIntegracion     varchar(50)   = ?
declare @CodigoCliente         varchar(100)  = ?
declare @NumeroDocumento       varchar(20)   = ?
declare @CUF                   varchar(100)  = ?
declare @EstadoDocumentoFiscal varchar(255)  = ?
declare @CodigoSucursalSIN     int           = ?
declare @TipoEmision           int           = ?
declare @IdSucursal            int           = ?
declare @FechaDesde            datetime2     = ?
declare @FechaHasta            datetime2     = ?

SELECT
    sdf.id                              AS id_documento_fiscal,
    sdf.numero_factura,
    sdf.codigo_integracion,
    sdf.tipo_factura,
    sdf.tipo_emision,
    sdf.estado_documento_fiscal,
    sdf.codigos_errores_sin,
    sdf.codigo_respuesta_sin,
    sdf.codigo_recepcion_sin,
    sdf.fecha_emision,
    sdf.fecha_envio,
    sdf.created_date,
    sdf.last_modified_date,
    sdf.cuf,
    sdf.cufd,
    sdf.cuis,
    sdf.nombre_razon_social,
    sdf.numero_documento,
    sdf.codigo_cliente,
    sdf.correo_electronico_cliente,
    sdf.monto_total,
    sdf.monto_total_moneda,
    sdf.usuario_emision,
    ss.id                               AS id_sucursal,
    ss.codigo_sucursal,
    ss.codigo_sucursal_sin,
    ss.nombre                           AS nombre_sucursal,
    ss.estado_sucursal,
    sp.id                               AS id_paquete,
    sp.estado_paquete,
    sp.cufd                             AS cufd_paquete,
    sp.cuis                             AS cuis_paquete,
    sp.codigo_respuesta_sin             AS codigo_respuesta_sin_paquete,
    sp.codigos_errores_sin              AS errores_sin_paquete,
    sp.count_invoices                   AS cant_facturas_paquete,
    sp.fecha_envio                      AS fecha_envio_paquete,
    se.id                               AS id_evento,
    se.evento                           AS nombre_evento,
    se.tipo_evento,
    se.state_evento,
    se.fecha_inicio                     AS evento_fecha_inicio,
    se.fecha_fin                        AS evento_fecha_fin,
    sddf.cantidad,
    sddf.codigo_producto_sfe,
    sddf.descripcion,
    sddf.sub_total,
    au.username                         AS usuario_creador

FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf WITH (NOLOCK)
JOIN FacturacionNaabol.dbo.sfe_detalle_documento_fiscal sddf
    ON sddf.id_sfe_documento_fiscal = sdf.id
JOIN FacturacionNaabol.dbo.sfe_sucursal ss
    ON ss.id = sdf.id_sfe_sucursal
LEFT JOIN FacturacionNaabol.dbo.sfe_paquete sp
    ON sp.id = sdf.id_paquete
LEFT JOIN FacturacionNaabol.dbo.sfe_evento se
    ON se.id = sdf.id_evento
LEFT JOIN FacturacionNaabol.dbo.auth_usuario au
    ON au.id = sdf.created_by

WHERE sdf.created_date >= @FechaDesde
  AND sdf.created_date <  @FechaHasta
  AND (@NumeroFactura         IS NULL OR sdf.numero_factura        = @NumeroFactura)
  AND (@CodigoIntegracion     IS NULL OR sdf.codigo_integracion    = @CodigoIntegracion)
  AND (@CodigoCliente         IS NULL OR sdf.codigo_cliente        = @CodigoCliente)
  AND (@NumeroDocumento       IS NULL OR sdf.numero_documento      = @NumeroDocumento)
  AND (@CUF                   IS NULL OR sdf.cuf                   = @CUF)
  AND (@EstadoDocumentoFiscal IS NULL OR sdf.estado_documento_fiscal = @EstadoDocumentoFiscal)
  AND (@CodigoSucursalSIN     IS NULL OR ss.codigo_sucursal_sin    = @CodigoSucursalSIN)
  AND (@TipoEmision           IS NULL OR sdf.tipo_emision          = @TipoEmision)
  AND (@IdSucursal            IS NULL OR ss.id                     = @IdSucursal)
  {{CODIGO_PRODUCTO_FILTER}}

ORDER BY sdf.created_date DESC;
`

func (s *ConsultasService) DataFacturas(data models.Json_consulta_data) (*[]models.SFEReporteFacturador, error) {

	idServer, err := strconv.ParseInt(data.IdFacturador, 10, 64)
	if err != nil {
		return nil, err
	}
	server, errServer := s.ConsultasRepo.GetServidorById(idServer)
	if errServer != nil {
		return nil, errServer
	}

	dsn := server.DSN()

	var db *gorm.DB

	db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("error al crear nueva conexión: %w", err)
		}
	}

	fechaDesde, err := time.Parse("2006-01-02", data.FechaDesde)
	if err != nil {
		return nil, fmt.Errorf("fechaDesde inválida: %w", err)
	}
	fechaHasta, err := time.Parse("2006-01-02", data.FechaHasta)
	if err != nil {
		return nil, fmt.Errorf("fechaHasta inválida: %w", err)
	}

	fechaHastaExclusiva := fechaHasta.AddDate(0, 0, 1)

	var numeroFactura any
	if data.NumeroFactura != "" {
		nf, errNf := strconv.ParseFloat(data.NumeroFactura, 64)
		if errNf != nil {
			return nil, fmt.Errorf("numeroFactura inválido: %w", errNf)
		}
		numeroFactura = nf
	}

	var tipoEmision any
	if data.TipoEmision != "" {
		te, errTe := strconv.Atoi(data.TipoEmision)
		if errTe != nil {
			return nil, fmt.Errorf("tipoEmision inválido: %w", errTe)
		}
		tipoEmision = te
	}

	var codigoSucursalSin any
	if data.CodigoSucursalSin != "" {
		cs, errCs := strconv.Atoi(data.CodigoSucursalSin)
		if errCs != nil {
			return nil, fmt.Errorf("codigoSucursalSin inválido: %w", errCs)
		}
		codigoSucursalSin = cs
	}

	var idSucursal any
	if data.Sucursal != "" {
		sid, errSid := strconv.Atoi(data.Sucursal)
		if errSid != nil {
			return nil, fmt.Errorf("sucursal inválida: %w", errSid)
		}
		idSucursal = sid
	}

	query := dataFacturasQuery
	args := []any{
		numeroFactura,
		toNullString(data.CodigoIntegracion),
		toNullString(data.CodigoCliente),
		toNullString(data.NumeroDocumento),
		toNullString(data.CUF),
		toNullString(data.EstadoDocumentoFiscal),
		codigoSucursalSin,
		tipoEmision,
		idSucursal,
		fechaDesde,
		fechaHastaExclusiva,
	}
	if len(data.CodigoProducto) > 0 {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(data.CodigoProducto)), ",")
		query = strings.Replace(query, "{{CODIGO_PRODUCTO_FILTER}}",
			"AND sddf.codigo_producto_sfe IN ("+placeholders+")", 1)
		for _, codigo := range data.CodigoProducto {
			args = append(args, codigo)
		}
	} else {
		query = strings.Replace(query, "{{CODIGO_PRODUCTO_FILTER}}", "", 1)
	}

	var facturas []models.SFEReporteFacturador
	if err := db.Raw(query, args...).Scan(&facturas).Error; err != nil {
		return nil, fmt.Errorf("error al buscar facturas: %w", err)
	}

	return &facturas, nil
}

func (s *ConsultasService) BuscarDuas(idServer string, params models.DuasBusquedaParams) (*models.DuasResultado, error) {
	idServerParse, err := strconv.ParseInt(idServer, 10, 64)
	if err != nil {
		return nil, err
	}

	Server, errServer := s.ConsultasRepo.GetServidorById(idServerParse)
	if errServer != nil {
		return nil, errServer
	}

	dsn := Server.DSN()

	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("error de conectividad: %w", err)
	}

	var Resultados models.DuasResultado
	var ResultadosCentral []models.DuasResultadoCentral
	var ResultadosLocal []models.DuasResultadoLocal

	toNull := toNullString
	var Query string
	if Server.Type == "duas" {
		Query = `
		declare @etkt varchar(50) = ?
		declare @nombre varchar(100) = ?
		declare @apellido varchar(50)= ?
		declare @vuelo varchar(10)=?
		declare @asiento varchar(5)= ?
		declare @origen varchar(3)= ?
		declare @fechaA varchar(50)=?
		declare @fechaB varchar(50)=?

		 select tf.IDTES_FACTURA_ITINERARIO ,tf.FAC_NROVUELO ,tf.FAC_FECHAEMISION_FACTURA ,tf.FAC_DETALLEFACTURA ,tf.FAC_MONTO ,tf.FECHACREACION,tf.IDTES_FACTURAONLINE  ,tf.IDTES_FACTURAONLINE as "ID_DOCUMENTO" ,tf.FAC_NROFACTURA as "NUMEROFACTURA" ,tf.FAC_FECHAEMISION_FACTURA as "FECHA_EMISION"   ,tf.URL_SIN as "URL_SIN" 
                from TES_FACTURAITINERARIO tf 
                where (@nombre is null or tf.FAC_DETALLEFACTURA like '%' + @nombre + '%')
                and (@etkt is null or tf.FAC_DETALLEFACTURA like '%2A' + @etkt + '%') 
                and (@apellido is null or tf.FAC_DETALLEFACTURA like '%M1' + @apellido + '%')
                and (@vuelo is null or tf.FAC_DETALLEFACTURA like '%' + @vuelo + '%')
                and (@asiento is null or tf.FAC_DETALLEFACTURA like '%' + @asiento + '%')
                and (@origen is null or tf.FAC_DETALLEFACTURA like '% ' + @origen + '%')
                and tf.FAC_FECHAEMISION_FACTURA between @fechaA and @fechaB;
		`
		if err := db.Raw(Query,
			toNull(params.Ticket),
			toNull(params.Nombre),
			toNull(params.Apellido),
			toNull(params.NumeroVuelo),
			toNull(params.Asiento),
			nil, 
			toNull(params.FechaDesde),
			toNull(params.FechaHasta),
		).Scan(&ResultadosCentral).Error; err != nil {
			return nil, fmt.Errorf("error ejecutando consulta DUAS Central: %w", err)
		}
	}
	if Server.Type == "duas_local" {
		Query = `
		declare @etkt varchar(50) = ?
		declare @nombre varchar(100) = ?
		declare @apellido varchar(50)= ?
		declare @vuelo varchar(10)=?
		declare @asiento varchar(5)= ?
		declare @origen varchar(3)= ?
		declare @fechaA varchar(50)=?
		declare @fechaB varchar(50)=?

		select tf.IDTES_FACTURA_ITINERARIO ,tf.FAC_NROVUELO ,tf.FAC_FECHAEMISION_FACTURA ,tf.FAC_DETALLEFACTURA ,tf.FAC_MONTO ,tf.FECHACREACION,tf.IDTES_FACTURAONLINE ,tf2.CODIGO ,tf2.ID_DOCUMENTO ,tf2.CUF ,tf2.CUFD ,tf2.CUIS ,tf2.NUMEROFACTURA ,tf2.FECHA_EMISION ,tf2.ESTADO_DOCUMENTO_FISCAL ,tf2.CODIGO_INTEGRACION ,tf2.URL_SIN 
		from TES_FACTURAITINERARIO tf 
		join TES_FACTURAONLINE tf2 on tf.IDTES_FACTURAONLINE = tf2.ID_DOCUMENTO 
		where (@nombre is null or tf.FAC_DETALLEFACTURA like '%' + @nombre + '%')
		and (@etkt is null or tf.FAC_DETALLEFACTURA like '%2A' + @etkt + '%') 
		and (@apellido is null or tf.FAC_DETALLEFACTURA like '%M1' + @apellido + '%')
		and (@vuelo is null or tf.FAC_DETALLEFACTURA like '%' + @vuelo + '%')
		and (@asiento is null or tf.FAC_DETALLEFACTURA like '%' + @asiento + '%')
		and (@origen is null or tf.FAC_DETALLEFACTURA like '% ' + @origen + '%')
		and tf.FAC_FECHAEMISION_FACTURA between @fechaA and @fechaB;
		`
		if err := db.Raw(Query,
			toNull(params.Ticket),
			toNull(params.Nombre),
			toNull(params.Apellido),
			toNull(params.NumeroVuelo),
			toNull(params.Asiento),
			nil, 
			toNull(params.FechaDesde),
			toNull(params.FechaHasta),
		).Scan(&ResultadosLocal).Error; err != nil {
			return nil, fmt.Errorf("error ejecutando consulta DUAS Local: %w", err)
		}
	}

	Resultados.ResultadosCentral = ResultadosCentral
	Resultados.ResultadosLocal = ResultadosLocal

	return &Resultados, nil
}

func sqlSubstring(s string, start, length int) string {
	r := []rune(s)
	if start < 1 {
		length += start - 1
		start = 1
	}
	if length <= 0 || start > len(r) {
		return ""
	}
	from := start - 1
	to := min(from+length, len(r))
	return string(r[from:to])
}

func (s *ConsultasService) Sucursales(idServer string) (*[]models.SFE_sucursales, error) {
	idServer_parse, err := strconv.ParseInt(idServer, 10, 64)
	if err != nil {
		return nil, err
	}

	if cache, errCache := s.SucursalesCache.ListByConexion(uint(idServer_parse)); errCache != nil {
		fmt.Printf("Error leyendo copia local de sucursales (conexión %d), se consulta en vivo: %v\n", idServer_parse, errCache)
	} else if len(cache) > 0 {
		sucursales := aSFESucursales(cache)
		return &sucursales, nil
	}

	server, errServer := s.ConsultasRepo.GetServidorById(idServer_parse)
	if errServer != nil {
		return nil, errServer
	}

	dsn := server.DSN()

	var db *gorm.DB

	db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("error al crear nueva conexión: %w", err)
		}
	}
	var servidores []models.SFE_sucursales
	errDataSuc := db.Table("sfe_sucursal").Find(&servidores).Error
	if errDataSuc != nil {
		return nil, errDataSuc
	}
	return &servidores, nil
}

const facturasMesQuery = `
declare @Producto    varchar(20) = ?
declare @IdSucursal  int         = ?
declare @CodigoSin   int         = ?
declare @FechaDesde  datetime2   = ?
declare @FechaHasta  datetime2   = ?

SELECT sdf.numero_factura, sdf.fecha_emision, sdf.cuf, sdf.cufd, sdf.monto_total,
       sddf.codigo_producto_sfe, sddf.descripcion
FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf WITH (NOLOCK)
JOIN FacturacionNaabol.dbo.sfe_detalle_documento_fiscal sddf WITH (NOLOCK) ON sddf.id_sfe_documento_fiscal = sdf.id
JOIN FacturacionNaabol.dbo.sfe_sucursal ss WITH (NOLOCK) ON ss.id = sdf.id_sfe_sucursal
WHERE sddf.codigo_producto_sfe = @Producto
  AND sdf.id_sfe_sucursal = @IdSucursal
  AND ss.codigo_sucursal_sin = @CodigoSin
  AND sdf.estado_documento_fiscal = 'VERIFICADO'
  AND sdf.fecha_emision >= @FechaDesde AND sdf.fecha_emision < @FechaHasta
ORDER BY sdf.fecha_emision ASC;
`

func (s *ConsultasService) FacturasMes(idServer int64, idSucursal, codigoSin int, producto string, anio, mes int) ([]models.FacturaMensual, error) {
	server, err := s.ConsultasRepo.GetServidorById(idServer)
	if err != nil {
		return nil, err
	}
	dsn := server.DSN()

	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	desde := time.Date(anio, time.Month(mes), 1, 0, 0, 0, 0, time.UTC)
	hasta := desde.AddDate(0, 1, 0)

	var facturas []models.FacturaMensual
	if err := db.Raw(facturasMesQuery, producto, idSucursal, codigoSin, desde, hasta).Scan(&facturas).Error; err != nil {
		return nil, fmt.Errorf("error al buscar facturas del mes: %w", err)
	}
	return facturas, nil
}
