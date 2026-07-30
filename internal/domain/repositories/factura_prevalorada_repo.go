package repositories

import (
	"errors"
	"fmt"
	"managerfact/internal/domain/models"
	"time"

	"gorm.io/gorm"
)

// LoteResumen agrega las facturas prevaloradas de un mismo lote de
// importación, con el contexto con el que se cargó (sucursal, tipo) y el
// desglose de estados de envío.
type LoteResumen struct {
	LoteID                   string `json:"lote_id"`
	SucursalFacturadorID     uint   `json:"sucursal_facturador_id"`
	SucursalFacturadorNombre string `json:"sucursal_facturador_nombre"`
	// CodigoSucursalSin no se serializa: solo se usa para filtrar el lote por
	// las sucursales permitidas del usuario (ver
	// FacturaPrevaloradaService.ListarLotes).
	CodigoSucursalSin int       `json:"-"`
	Tipo              string    `json:"tipo"`
	Observacion       string    `json:"observacion"`
	Total             int64     `json:"total"`
	Pendientes        int64     `json:"pendientes"`
	Enviados          int64     `json:"enviados"`
	Aceptados         int64     `json:"aceptados"`
	Rechazados        int64     `json:"rechazados"`
	ConError          int64     `json:"con_error"`
	FechaImportacion  time.Time `json:"fecha_importacion"`
}

type FacturaPrevaloradaRepository struct {
	db *gorm.DB
}

func NewFacturaPrevaloradaRepository(db *gorm.DB) *FacturaPrevaloradaRepository {
	return &FacturaPrevaloradaRepository{db: db}
}

func (r *FacturaPrevaloradaRepository) Create(factura *models.FacturaPrevalorada) error {
	if err := r.db.Create(factura).Error; err != nil {
		return fmt.Errorf("error creando factura prevalorada: %w", err)
	}
	return nil
}

// CreateBatch inserta todas las filas válidas de una importación en una sola
// transacción.
func (r *FacturaPrevaloradaRepository) CreateBatch(facturas []models.FacturaPrevalorada) error {
	if len(facturas) == 0 {
		return nil
	}
	if err := r.db.Create(&facturas).Error; err != nil {
		return fmt.Errorf("error guardando facturas prevaloradas: %w", err)
	}
	return nil
}

// Update guarda el resultado del envío al facturador (etapa 2): solo toca
// las columnas de seguimiento, nunca los datos importados en la etapa 1 ni
// la asociación SucursalFacturador.
func (r *FacturaPrevaloradaRepository) Update(factura *models.FacturaPrevalorada) error {
	err := r.db.Model(&models.FacturaPrevalorada{}).Where("id = ?", factura.ID).Updates(map[string]interface{}{
		"estado":            factura.Estado,
		"codigo_respuesta":  factura.CodigoRespuesta,
		"mensaje_respuesta": factura.MensajeRespuesta,
		"fecha_envio":       factura.FechaEnvio,
		"fecha_respuesta":   factura.FechaRespuesta,
		"cuf":               factura.CUF,
		"numero_factura":    factura.NumeroFactura,
		"url_documento":     factura.UrlDocumento,
	}).Error
	if err != nil {
		return fmt.Errorf("error actualizando factura prevalorada: %w", err)
	}
	return nil
}

func (r *FacturaPrevaloradaRepository) GetByID(id uint) (*models.FacturaPrevalorada, error) {
	var factura models.FacturaPrevalorada
	err := r.db.Preload("SucursalFacturador").First(&factura, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("factura prevalorada con ID %d no encontrada", id)
		}
		return nil, fmt.Errorf("error obteniendo factura prevalorada: %w", err)
	}
	return &factura, nil
}

// GetAll lista facturas prevaloradas, filtrando opcionalmente por estado y/o
// lote_id (ambos vacíos = sin filtro).
func (r *FacturaPrevaloradaRepository) GetAll(estado, loteID string) ([]models.FacturaPrevalorada, error) {
	facturas := []models.FacturaPrevalorada{}
	query := r.db.Preload("SucursalFacturador")
	if estado != "" {
		query = query.Where("estado = ?", estado)
	}
	if loteID != "" {
		query = query.Where("lote_id = ?", loteID)
	}
	if err := query.Order("created_at DESC").Find(&facturas).Error; err != nil {
		return nil, fmt.Errorf("error obteniendo facturas prevaloradas: %w", err)
	}
	return facturas, nil
}

// GetPendientesParaEnvio lista las facturas pendientes en el orden en que el
// EnvioWorker debe procesarlas: agrupadas por sucursal (para poder aplicar
// el circuit breaker por sucursal) y luego por lote/orden de creación.
func (r *FacturaPrevaloradaRepository) GetPendientesParaEnvio() ([]models.FacturaPrevalorada, error) {
	facturas := []models.FacturaPrevalorada{}
	err := r.db.Preload("SucursalFacturador").
		Where("estado = ?", "pendiente").
		Order("sucursal_facturador_id ASC, lote_id ASC, created_at ASC").
		Find(&facturas).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo facturas pendientes: %w", err)
	}
	return facturas, nil
}

// GetLotes agrega las facturas prevaloradas por lote_id: sucursal/tipo con
// los que se cargó el lote, total de filas y desglose por estado.
func (r *FacturaPrevaloradaRepository) GetLotes() ([]LoteResumen, error) {
	lotes := []LoteResumen{}
	err := r.db.Table("facturas_prevaloradas AS fp").
		Select(`
			fp.lote_id,
			fp.sucursal_facturador_id,
			sf.nombre AS sucursal_facturador_nombre,
			sf.codigo_sucursal_sin,
			fp.tipo,
			MIN(fp.observacion) AS observacion,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE fp.estado = 'pendiente') AS pendientes,
			COUNT(*) FILTER (WHERE fp.estado = 'enviado') AS enviados,
			COUNT(*) FILTER (WHERE fp.estado = 'aceptado') AS aceptados,
			COUNT(*) FILTER (WHERE fp.estado = 'rechazado') AS rechazados,
			COUNT(*) FILTER (WHERE fp.estado = 'error') AS con_error,
			MIN(fp.created_at) AS fecha_importacion
		`).
		Joins("JOIN sucursales_facturador AS sf ON sf.id = fp.sucursal_facturador_id").
		Where("fp.deleted_at IS NULL").
		Group("fp.lote_id, fp.sucursal_facturador_id, sf.nombre, sf.codigo_sucursal_sin, fp.tipo").
		Order("MIN(fp.created_at) DESC").
		Scan(&lotes).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo lotes: %w", err)
	}
	return lotes, nil
}
