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

// toNullString convierte un string vacío en NULL para respetar la semántica
// "@param IS NULL" del query (usado en todas las consultas parametrizadas).
func toNullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

type ConsultasService struct {
	ConsultasRepo repositories.ConsutasRepository
}

func NewConsultasService(r *repositories.ConsutasRepository) *ConsultasService {
	return &ConsultasService{
		ConsultasRepo: *r,
	}
}

// dataFacturasQuery es el reporte completo de facturación: documento fiscal +
// detalle + sucursal + paquete (offline/contingencia) + evento + usuario que
// registró el documento. Todos los filtros son opcionales salvo el rango de
// fechas, que se aplica sobre created_date (fecha de registro del documento).
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

FROM FacturacionNaabol.dbo.sfe_documento_fiscal sdf
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
	// Construcción del DSN (Data Source Name) para SQL Server
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		server.Username,
		server.Password,
		server.Host,
		server.Port,
		server.DatabaseName,
	)

	var db *gorm.DB

	// Intentar conectar primero para verificar si hay conexión activa
	db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar: %w", err)
	}

	// Verificar si la conexión está activa
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	// Ping para confirmar conexión
	if err := sqlDB.Ping(); err != nil {
		// Si falla, intentar crear nueva conexión
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
	// Límite superior exclusivo (inicio del día siguiente).
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

	// El filtro de código de producto admite selección múltiple, por lo que se
	// arma un IN (...) con tantos placeholders como códigos se hayan enviado.
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

	if len(facturas) == 0 {
		return nil, fmt.Errorf("no se encontraron facturas")
	}

	return &facturas, nil
}

// BuscarDuas ejecuta la consulta DUAS contra la base de datos seleccionada
func (s *ConsultasService) BuscarDuas(idServer string, params models.DuasBusquedaParams) (*models.DuasResultado, error) {
	idServerParse, err := strconv.ParseInt(idServer, 10, 64)
	if err != nil {
		return nil, err
	}

	Server, errServer := s.ConsultasRepo.GetServidorById(idServerParse)
	if errServer != nil {
		return nil, errServer
	}

	// Construcción del DSN para SQL Server
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		Server.Username,
		Server.Password,
		Server.Host,
		Server.Port,
		Server.DatabaseName,
	)

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

	// toNull convierte un string vacío en NULL para respetar la semántica
	// "@param IS NULL" del query (clave para los parámetros DATE: '' no es
	// convertible a DATE en SQL Server y rompería la consulta).
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
			nil, // @origen: sin parámetro de búsqueda por ahora
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
			nil, // @origen: sin parámetro de búsqueda por ahora
			toNull(params.FechaDesde),
			toNull(params.FechaHasta),
		).Scan(&ResultadosLocal).Error; err != nil {
			return nil, fmt.Errorf("error ejecutando consulta DUAS Local: %w", err)
		}
	}
	// 	if Server.Type == "duas_central" {
	// 		Query = `
	// 			DECLARE @nombre      NVARCHAR(50) = ?
	// 			DECLARE @apellido    NVARCHAR(50) = ?
	// 			DECLARE @vuelo       NVARCHAR(10) = ?
	// 			DECLARE @fecha_desde DATE         = ?
	// 			DECLARE @fecha_hasta DATE         = ?
	// 			DECLARE @asiento     NVARCHAR(10) = ?
	// 			DECLARE @ticket      NVARCHAR(13) = ?

	// 			SELECT
	// 				t.IDTES_FACTURA_ITINERARIO,
	// 				t.FAC_NROFACTURA,
	// 				t.FAC_NROVUELO,
	// 				t.FAC_FECHAHORA_VUELO,
	// 				t.FAC_MONTO,
	// 				t.IDA_ESTADOFACTURA,
	// 				t.FECHACREACION,
	// 				t.USUARIOCREACION,
	// 				t.URL_SIN,
	// 				t.FAC_DETALLEFACTURA,
	// 				t.FAC_FECHAEMISION_FACTURA

	// 			FROM TES_FACTURAITINERARIO t
	// 			CROSS APPLY (
	// 				SELECT
	// 					UPPER(SUBSTRING(t.FAC_DETALLEFACTURA, 3, 20)) AS bloqueNombre
	// 			) p
	// 			WHERE
	// 				(@apellido IS NULL OR p.bloqueNombre LIKE '%' + UPPER(@apellido) + '%')
	// 				AND (@nombre   IS NULL OR p.bloqueNombre LIKE '%' + UPPER(@nombre)   + '%')
	// 				AND (@vuelo    IS NULL OR t.FAC_NROVUELO = @vuelo)
	// 				AND (
	// 					@fecha_desde IS NULL
	// 					OR (@fecha_hasta IS NULL     AND CAST(t.FAC_FECHAHORA_VUELO AS DATE) = @fecha_desde)
	// 					OR (@fecha_hasta IS NOT NULL AND CAST(t.FAC_FECHAHORA_VUELO AS DATE) BETWEEN @fecha_desde AND @fecha_hasta)
	// 				)
	// 				AND (@asiento IS NULL OR t.FAC_DETALLEFACTURA LIKE '%' + @asiento + '%')
	// 				AND (@ticket  IS NULL OR
	// 					CASE WHEN CHARINDEX('2A', t.FAC_DETALLEFACTURA) > 0
	// 						THEN SUBSTRING(t.FAC_DETALLEFACTURA, CHARINDEX('2A', t.FAC_DETALLEFACTURA) + 2, 13)
	// 						ELSE NULL END = @ticket
	// 				)

	// 			ORDER BY t.FAC_FECHAHORA_VUELO DESC
	// `
	// 		if err := db.Raw(Query,
	// 			toNull(params.Nombre),
	// 			toNull(params.Apellido),
	// 			toNull(params.NumeroVuelo),
	// 			toNull(params.FechaDesde),
	// 			toNull(params.FechaHasta),
	// 			toNull(params.Asiento),
	// 			toNull(params.Ticket),
	// 		).Scan(&ResultadosCentral).Error; err != nil {
	// 			return nil, fmt.Errorf("error ejecutando consulta DUAS Central: %w", err)
	// 		}
	// 	}
	Resultados.ResultadosCentral = ResultadosCentral
	Resultados.ResultadosLocal = ResultadosLocal

	// Parsear el BCBP embebido en FAC_DETALLEFACTURA para poblar los campos
	// que la tabla del front necesita (apellido, nombre, origen, etc.).

	return &Resultados, nil
}

// sqlSubstring replica el comportamiento de SUBSTRING de SQL Server:
// posición inicial 1-based y longitud recortada al final de la cadena.
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

// parseBCBP descompone FAC_DETALLEFACTURA (boarding pass BCBP) en los campos
// del resultado, equivalente a las expresiones SUBSTRING/CHARINDEX del query.

func (s *ConsultasService) Sucursales(idServer string) (*[]models.SFE_sucursales, error) {
	idServer_parse, err := strconv.ParseInt(idServer, 10, 64)
	if err != nil {
		return nil, err
	}
	server, errServer := s.ConsultasRepo.GetServidorById(idServer_parse)
	if errServer != nil {
		return nil, errServer
	}
	// Construcción del DSN (Data Source Name) para SQL Server
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		server.Username,
		server.Password,
		server.Host,
		server.Port,
		server.DatabaseName,
	)

	var db *gorm.DB

	// Intentar conectar primero para verificar si hay conexión activa
	db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar: %w", err)
	}

	// Verificar si la conexión está activa
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	// Ping para confirmar conexión
	if err := sqlDB.Ping(); err != nil {
		// Si falla, intentar crear nueva conexión
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
