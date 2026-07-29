package models

import (
	"time"

	"gorm.io/gorm"
)

// FacturaAnulacion es una solicitud de anulación de una factura ya emitida,
// importada por lote desde Excel — ver doc/EnvioFacturacion.md sección 4.
// A diferencia de FacturaPrevalorada no describe un ítem nuevo a facturar,
// solo referencia una factura existente (por CUF + CodigoIntegracion) y el
// motivo de anulación, por eso vive en su propia tabla en vez de columnas
// nullable sobre facturas_prevaloradas.
type FacturaAnulacion struct {
	ID                   uint                `json:"id" gorm:"primaryKey"`
	SucursalFacturadorID uint                `json:"sucursal_facturador_id" gorm:"not null"`
	SucursalFacturador   *SucursalFacturador `json:"sucursal_facturador,omitempty" gorm:"foreignKey:SucursalFacturadorID"`
	LoteID               string              `json:"lote_id" gorm:"type:varchar(36);not null;index"`
	// Observacion es el motivo de carga del lote (por qué se importó este
	// archivo), ingresado una sola vez antes de importar y fijo para todas
	// las filas del lote — mismo patrón que en FacturaPrevalorada.
	Observacion string `json:"observacion" gorm:"type:varchar(255);not null"`

	// Etapa 1: datos importados del Excel. A diferencia de
	// FacturaPrevalorada, CodigoIntegracion NO se genera acá: es el de la
	// factura original que se quiere anular, viene del Excel junto con Cuf.
	CodigoIntegracion string `json:"codigo_integracion" gorm:"type:varchar(64);not null;uniqueIndex"`
	Cuf               string `json:"cuf" gorm:"type:varchar(250);not null"`
	CodigoMotivo      string `json:"codigo_motivo" gorm:"type:varchar(10);not null"`

	// Etapa 2: seguimiento de envío al facturador.
	Estado           string     `json:"estado" gorm:"type:varchar(20);not null;default:'pendiente';index"`
	CodigoRespuesta  string     `json:"codigo_respuesta" gorm:"type:varchar(50)"`
	MensajeRespuesta string     `json:"mensaje_respuesta" gorm:"type:text"`
	FechaEnvio       *time.Time `json:"fecha_envio"`
	FechaRespuesta   *time.Time `json:"fecha_respuesta"`
	IntentosConsulta int        `json:"intentos_consulta" gorm:"default:0"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (FacturaAnulacion) TableName() string { return "facturas_anulacion" }
