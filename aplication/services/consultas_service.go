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
	query = query.Where("df.fecha_emision >= ? AND df.fecha_emision <= ?", data.FechaDesde, data.FechaHasta)
	if data.Sucursal != "" {
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
