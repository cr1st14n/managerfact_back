package repositories

import (
	"errors"
	"fmt"
	"managerfact/internal/domain/models"
	"time"

	"gorm.io/gorm"
)

// LoteResumenAnulacion agrega las facturas de anulación de un mismo lote de
// importación, con el contexto con el que se cargó (sucursal, observación) y
// el desglose de estados de envío.
type LoteResumenAnulacion struct {
	LoteID                   string    `json:"lote_id"`
	SucursalFacturadorID     uint      `json:"sucursal_facturador_id"`
	SucursalFacturadorNombre string    `json:"sucursal_facturador_nombre"`
	Observacion              string    `json:"observacion"`
	Total                    int64     `json:"total"`
	Pendientes               int64     `json:"pendientes"`
	Enviados                 int64     `json:"enviados"`
	Aceptados                int64     `json:"aceptados"`
	Rechazados               int64     `json:"rechazados"`
	ConError                 int64     `json:"con_error"`
	FechaImportacion         time.Time `json:"fecha_importacion"`
}

type FacturaAnulacionRepository struct {
	db *gorm.DB
}

func NewFacturaAnulacionRepository(db *gorm.DB) *FacturaAnulacionRepository {
	return &FacturaAnulacionRepository{db: db}
}

func (r *FacturaAnulacionRepository) Create(factura *models.FacturaAnulacion) error {
	if err := r.db.Create(factura).Error; err != nil {
		return fmt.Errorf("error creando factura de anulación: %w", err)
	}
	return nil
}

// CreateBatch inserta todas las filas válidas de una importación en una sola
// transacción.
func (r *FacturaAnulacionRepository) CreateBatch(facturas []models.FacturaAnulacion) error {
	if len(facturas) == 0 {
		return nil
	}
	if err := r.db.Create(&facturas).Error; err != nil {
		return fmt.Errorf("error guardando facturas de anulación: %w", err)
	}
	return nil
}

func (r *FacturaAnulacionRepository) GetByID(id uint) (*models.FacturaAnulacion, error) {
	var factura models.FacturaAnulacion
	err := r.db.Preload("SucursalFacturador").First(&factura, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("factura de anulación con ID %d no encontrada", id)
		}
		return nil, fmt.Errorf("error obteniendo factura de anulación: %w", err)
	}
	return &factura, nil
}

// GetAll lista facturas de anulación, filtrando opcionalmente por estado y/o
// lote_id (ambos vacíos = sin filtro).
func (r *FacturaAnulacionRepository) GetAll(estado, loteID string) ([]models.FacturaAnulacion, error) {
	facturas := []models.FacturaAnulacion{}
	query := r.db.Preload("SucursalFacturador")
	if estado != "" {
		query = query.Where("estado = ?", estado)
	}
	if loteID != "" {
		query = query.Where("lote_id = ?", loteID)
	}
	if err := query.Order("created_at DESC").Find(&facturas).Error; err != nil {
		return nil, fmt.Errorf("error obteniendo facturas de anulación: %w", err)
	}
	return facturas, nil
}

// GetLotes agrega las facturas de anulación por lote_id: sucursal y
// observación con las que se cargó el lote, total de filas y desglose por
// estado.
func (r *FacturaAnulacionRepository) GetLotes() ([]LoteResumenAnulacion, error) {
	lotes := []LoteResumenAnulacion{}
	err := r.db.Table("facturas_anulacion AS fa").
		Select(`
			fa.lote_id,
			fa.sucursal_facturador_id,
			sf.nombre AS sucursal_facturador_nombre,
			MIN(fa.observacion) AS observacion,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE fa.estado = 'pendiente') AS pendientes,
			COUNT(*) FILTER (WHERE fa.estado = 'enviado') AS enviados,
			COUNT(*) FILTER (WHERE fa.estado = 'aceptado') AS aceptados,
			COUNT(*) FILTER (WHERE fa.estado = 'rechazado') AS rechazados,
			COUNT(*) FILTER (WHERE fa.estado = 'error') AS con_error,
			MIN(fa.created_at) AS fecha_importacion
		`).
		Joins("JOIN sucursales_facturador AS sf ON sf.id = fa.sucursal_facturador_id").
		Where("fa.deleted_at IS NULL").
		Group("fa.lote_id, fa.sucursal_facturador_id, sf.nombre").
		Order("MIN(fa.created_at) DESC").
		Scan(&lotes).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo lotes: %w", err)
	}
	return lotes, nil
}
