package models

import "time"

type LogEnvio struct {
	ID                   uint                `json:"id" gorm:"primaryKey"`
	Tipo                 string              `json:"tipo" gorm:"type:varchar(20);not null;index"` 
	FacturaID            uint                `json:"factura_id" gorm:"not null;index"`
	CodigoIntegracion    string              `json:"codigo_integracion" gorm:"type:varchar(64)"`
	SucursalFacturadorID uint                `json:"sucursal_facturador_id" gorm:"not null;index"`
	SucursalFacturador   *SucursalFacturador `json:"sucursal_facturador,omitempty" gorm:"foreignKey:SucursalFacturadorID"`

	Origen string `json:"origen" gorm:"type:varchar(20);not null"` 

	Resultado string    `json:"resultado" gorm:"type:varchar(20);not null;index"`
	Mensaje   string    `json:"mensaje" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}

func (LogEnvio) TableName() string { return "logs_envio" }
