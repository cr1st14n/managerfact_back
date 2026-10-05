package services

import (
	"context"
	"errors"
	"fmt"
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var (

	ErrConexionInvalida = errors.New("conexión inválida")

	ErrOrigenSucursales = errors.New("no se pudo leer del origen")
)

// Mantener esta consulta al origen como SELECT fijo, sin parámetros del usuario.
const sucursalesOrigenQuery = `
SELECT id AS sfe_id, codigo_sucursal, codigo_sucursal_sin, nombre, direccion,
       municipio_departamento, estado_sucursal
FROM sfe_sucursal WITH (NOLOCK)`

type ResultadoSincronizacion struct {
	Total         int       `json:"total"`
	Nuevas        int       `json:"nuevas"`
	ActualizadoAt time.Time `json:"actualizado_at"`
}

type SucursalConexion struct {
	ConexionID            uint   `json:"conexion_id"`
	ConexionNombre        string `json:"conexion_nombre"`
	ConexionAmbiente      string `json:"conexion_ambiente"`
	Host                  string `json:"host"`
	DatabaseName          string `json:"database_name"`
	ID                    string `json:"id"`
	CodigoSucursal        string `json:"codigo_sucursal"`
	CodigoSucursalSin     string `json:"codigo_sucursal_sin"`
	Nombre                string `json:"nombre"`
	Direccion             string `json:"direccion"`
	MunicipioDepartamento string `json:"municipio_departamento"`
	EstadoSucursal        string `json:"estado_sucursal"`
}

type ConexionSinSucursales struct {
	ID               uint   `json:"id"`
	ConexionNombre   string `json:"conexion_nombre"`
	ConexionAmbiente string `json:"conexion_ambiente"`
}

type SucursalesTodas struct {
	Sucursales    []SucursalConexion      `json:"sucursales"`
	SinActualizar []ConexionSinSucursales `json:"sin_actualizar"`
}

type ConexionSucursalService struct {
	conexiones repositories.DbConnectionRepository
	repo       *repositories.ConexionSucursalRepo
}

func NewConexionSucursalService(c repositories.DbConnectionRepository, r *repositories.ConexionSucursalRepo) *ConexionSucursalService {
	return &ConexionSucursalService{conexiones: c, repo: r}
}

func esConexionFacturador(c *models.DbConnection) bool {
	return strings.Contains(strings.ToLower(c.Type), "facturador")
}

// Único punto que actualiza la copia local; Reportes solo la leen.
func (s *ConexionSucursalService) Sincronizar(conexionID uint) (*ResultadoSincronizacion, error) {
	conexion, err := s.conexiones.GetByID(conexionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConexionInvalida, err)
	}
	if !esConexionFacturador(conexion) {
		return nil, fmt.Errorf("%w: solo las conexiones de tipo facturador tienen sucursales para actualizar", ErrConexionInvalida)
	}

	filas, err := leerSucursalesOrigen(conexion)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOrigenSucursales, err)
	}

	ahora := time.Now()
	for i := range filas {
		filas[i].ConexionID = conexionID
		filas[i].ActualizadoAt = ahora
	}

	nuevas, err := s.repo.Upsert(conexionID, filas)
	if err != nil {
		return nil, fmt.Errorf("error guardando las sucursales: %w", err)
	}

	return &ResultadoSincronizacion{Total: len(filas), Nuevas: nuevas, ActualizadoAt: ahora}, nil
}

func (s *ConexionSucursalService) ListarTodas() (*SucursalesTodas, error) {
	conexiones, err := s.conexiones.GetAllActiveByType("facturador")
	if err != nil {
		return nil, err
	}

	ids := make([]uint, 0, len(conexiones))
	porID := make(map[uint]models.DbConnection, len(conexiones))
	for _, c := range conexiones {
		ids = append(ids, c.ID)
		porID[c.ID] = c
	}

	filas, err := s.repo.ListByConexiones(ids)
	if err != nil {
		return nil, err
	}

	resultado := &SucursalesTodas{
		Sucursales:    make([]SucursalConexion, 0, len(filas)),
		SinActualizar: []ConexionSinSucursales{},
	}
	conCopia := make(map[uint]bool, len(conexiones))
	for _, f := range filas {
		c := porID[f.ConexionID]
		conCopia[f.ConexionID] = true
		resultado.Sucursales = append(resultado.Sucursales, SucursalConexion{
			ConexionID:            f.ConexionID,
			ConexionNombre:        c.ServerName,
			ConexionAmbiente:      c.Ambiente,
			Host:                  c.Host,
			DatabaseName:          c.DatabaseName,
			ID:                    strconv.FormatInt(f.SfeID, 10),
			CodigoSucursal:        f.CodigoSucursal,
			CodigoSucursalSin:     f.CodigoSucursalSin,
			Nombre:                f.Nombre,
			Direccion:             f.Direccion,
			MunicipioDepartamento: f.MunicipioDepartamento,
			EstadoSucursal:        f.EstadoSucursal,
		})
	}
	for _, c := range conexiones {
		if !conCopia[c.ID] {
			resultado.SinActualizar = append(resultado.SinActualizar, ConexionSinSucursales{ID: c.ID, ConexionNombre: c.ServerName, ConexionAmbiente: c.Ambiente})
		}
	}
	return resultado, nil
}

func (s *ConexionSucursalService) Resumen() ([]repositories.ResumenSucursales, error) {
	return s.repo.Resumen()
}

func leerSucursalesOrigen(c *models.DbConnection) ([]models.ConexionSucursal, error) {
	db, err := gorm.Open(sqlserver.Open(c.DSN()), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var filas []models.ConexionSucursal
	if err := db.WithContext(ctx).Raw(sucursalesOrigenQuery).Scan(&filas).Error; err != nil {
		return nil, err
	}
	return filas, nil
}

func aSFESucursales(filas []models.ConexionSucursal) []models.SFE_sucursales {
	out := make([]models.SFE_sucursales, 0, len(filas))
	for _, f := range filas {
		out = append(out, models.SFE_sucursales{
			ID:                    strconv.FormatInt(f.SfeID, 10),
			CodigoSucursal:        f.CodigoSucursal,
			CodigoSucursal_sin:    f.CodigoSucursalSin,
			Nombre:                f.Nombre,
			Direccion:             f.Direccion,
			MunicipioDepartamento: f.MunicipioDepartamento,
			EstadoSucursal:        f.EstadoSucursal,
		})
	}
	return out
}
