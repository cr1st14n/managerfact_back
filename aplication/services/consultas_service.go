package services

import (
	"fmt"
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"
	"strconv"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type ConsultasService struct {
	ConsultasRepo repositories.ConsutasRepository
}

func NewConsultasService(r *repositories.ConsutasRepository) *ConsultasService {
	return &ConsultasService{
		ConsultasRepo: *r,
	}
}

func (s *ConsultasService) DataFacturas(data models.Json_consulta_data) (*[]models.SFE_factura, error) {
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
	var facturas []models.SFE_factura

	// Construir query base
	query := db.Table("sfe_documento_fiscal as df").
		Select("df.numero_factura, df.nombre_razon_social, df.numero_documento, dff.codigo_producto_sfe, dff.descripcion, dff.sub_total, df.cuf, df.fecha_emision, df.fecha_envio, df.estado_documento_fiscal, df.usuario_emision, su.nombre, su.codigo_sucursal_sin, df.tipo_factura").
		Joins("join FacturacionNaabol.dbo.sfe_detalle_documento_fiscal as dff ON df.id = dff.id_sfe_documento_fiscal").
		Joins("join FacturacionNaabol.dbo.sfe_sucursal as su ON df.id_sfe_sucursal = su.id")

	// Aplicar filtros obligatorios
	query = query.Where("df.fecha_emision >= ? AND df.fecha_emision <= ?", data.FechaDesde, data.FechaHasta+" 23:59:59")

	if data.Sucursal != "" {
		fmt.Println("buscando con sucursal ", data.Sucursal)
		query = query.Where("su.id = ?", data.Sucursal)
	}

	// Aplicar filtros opcionales solo si no están vacíos
	if data.NumeroDocumento != "" {
		query = query.Where("df.numero_documento = ?", data.NumeroDocumento)
	}
	if data.NumeroFactura != "" {
		query = query.Where("df.numero_factura = ?", data.NumeroFactura)
	}

	if data.CodigoProducto != "" {
		query = query.Where("dff.codigo_producto_sfe = ?", data.CodigoProducto)
	}

	// Ejecutar query
	if err := query.Find(&facturas).Error; err != nil {
		return nil, fmt.Errorf("error al buscar facturas: %w", err)
	}

	if len(facturas) == 0 {
		return nil, fmt.Errorf("no se encontraron facturas")
	}

	// Aquí ya tienes conexión GORM activa verificada
	// Puedes usar 'db' para tus operaciones GORM

	return &facturas, nil
}

// BuscarDuas ejecuta la consulta DUAS contra la base de datos seleccionada
func (s *ConsultasService) BuscarDuas(idServer string, params models.DuasBusquedaParams) (*[]models.DuasResultado, error) {
	idServerParse, err := strconv.ParseInt(idServer, 10, 64)
	if err != nil {
		return nil, err
	}

	server, errServer := s.ConsultasRepo.GetServidorById(idServerParse)
	if errServer != nil {
		return nil, errServer
	}

	// Construcción del DSN para SQL Server
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		server.Username,
		server.Password,
		server.Host,
		server.Port,
		server.DatabaseName,
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

	var resultados []models.DuasResultado

	// toNull convierte un string vacío en NULL para respetar la semántica
	// "@param IS NULL" del query (clave para los parámetros DATE: '' no es
	// convertible a DATE en SQL Server y rompería la consulta).
	toNull := func(s string) any {
		if s == "" {
			return nil
		}
		return s
	}

	query := `
DECLARE @nombre      NVARCHAR(50) = ?
DECLARE @apellido    NVARCHAR(50) = ?
DECLARE @vuelo       NVARCHAR(10) = ?
DECLARE @fecha_desde DATE         = ?
DECLARE @fecha_hasta DATE         = ?
DECLARE @asiento     NVARCHAR(10) = ?
DECLARE @ticket      NVARCHAR(13) = ?

SELECT
    t.IDTES_FACTURA_ITINERARIO,
    t.FAC_NROFACTURA,
    t.FAC_NROVUELO,
    t.FAC_FECHAHORA_VUELO,
    t.FAC_MONTO,
    t.IDA_ESTADOFACTURA,
    t.FECHACREACION,
    t.USUARIOCREACION,
    t.URL_SIN,
    t.FAC_DETALLEFACTURA,
	t.FAC_FECHAEMISION_FACTURA
FROM TES_FACTURAITINERARIO t
CROSS APPLY (
    SELECT
        CHARINDEX('#', t.FAC_DETALLEFACTURA)          AS posHash,
        CHARINDEX('#', t.FAC_DETALLEFACTURA) + 21     AS posBase
) p
WHERE
    p.posHash > 3
    AND (@apellido IS NULL OR
        SUBSTRING(t.FAC_DETALLEFACTURA, 3, p.posHash - 3) LIKE '%' + @apellido + '%'
    )
    AND (@nombre IS NULL OR
        LTRIM(RTRIM(SUBSTRING(t.FAC_DETALLEFACTURA, p.posHash + 1, 20))) LIKE @nombre + '%'
    )
    AND (@vuelo IS NULL OR t.FAC_NROVUELO = @vuelo)
    AND (
        @fecha_desde IS NULL
        OR (@fecha_hasta IS NULL     AND CAST(t.FAC_FECHAHORA_VUELO AS DATE) = @fecha_desde)
        OR (@fecha_hasta IS NOT NULL AND CAST(t.FAC_FECHAHORA_VUELO AS DATE) BETWEEN @fecha_desde AND @fecha_hasta)
    )
    AND (@asiento IS NULL OR t.FAC_DETALLEFACTURA LIKE '%' + @asiento + '%')
    AND (@ticket IS NULL OR
        CASE WHEN CHARINDEX('2A', t.FAC_DETALLEFACTURA) > 0
             THEN SUBSTRING(t.FAC_DETALLEFACTURA, CHARINDEX('2A', t.FAC_DETALLEFACTURA) + 2, 13)
             ELSE NULL END = @ticket
    )
ORDER BY t.FAC_FECHAHORA_VUELO DESC
`

	if err := db.Raw(query,
		toNull(params.Nombre),
		toNull(params.Apellido),
		toNull(params.NumeroVuelo),
		toNull(params.FechaDesde),
		toNull(params.FechaHasta),
		toNull(params.Asiento),
		toNull(params.Ticket),
	).Scan(&resultados).Error; err != nil {
		return nil, fmt.Errorf("error ejecutando consulta DUAS: %w", err)
	}

	// Parsear el BCBP embebido en FAC_DETALLEFACTURA para poblar los campos
	// que la tabla del front necesita (apellido, nombre, origen, etc.).

	return &resultados, nil
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
