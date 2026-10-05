package models

import (
	"time"

	"gorm.io/gorm"
)

type FacturaAnulacion struct {
	ID                   uint                `json:"id" gorm:"primaryKey"`
	SucursalFacturadorID uint                `json:"sucursal_facturador_id" gorm:"not null"`
	SucursalFacturador   *SucursalFacturador `json:"sucursal_facturador,omitempty" gorm:"foreignKey:SucursalFacturadorID"`
	LoteID               string              `json:"lote_id" gorm:"type:varchar(36);not null;index"`

	Observacion string `json:"observacion" gorm:"type:varchar(255);not null"`

	CodigoIntegracion string `json:"codigo_integracion" gorm:"type:varchar(64);not null;index"`
	Cuf               string `json:"cuf" gorm:"type:varchar(250);not null"`
	CodigoMotivo      string `json:"codigo_motivo" gorm:"type:varchar(10);not null"`

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
