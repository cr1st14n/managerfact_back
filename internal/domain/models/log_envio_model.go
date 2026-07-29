package models

import "time"

// LogEnvio registra cada intento de envío al facturador (prevalorada o
// anulación), sea disparado manualmente desde el botón del front o
// automáticamente por el EnvioWorker — permite ver en el front que el
// envío automático realmente está corriendo, sin tener que mirar los logs
// del servidor.
type LogEnvio struct {
	ID                   uint                `json:"id" gorm:"primaryKey"`
	Tipo                 string              `json:"tipo" gorm:"type:varchar(20);not null;index"` // "prevalorada" | "anulacion"
	FacturaID            uint                `json:"factura_id" gorm:"not null;index"`
	CodigoIntegracion    string              `json:"codigo_integracion" gorm:"type:varchar(64)"`
	SucursalFacturadorID uint                `json:"sucursal_facturador_id" gorm:"not null;index"`
	SucursalFacturador   *SucursalFacturador `json:"sucursal_facturador,omitempty" gorm:"foreignKey:SucursalFacturadorID"`
	// Origen distingue si el envío lo disparó el usuario (botón "Facturar"/
	// "Anular") o el EnvioWorker en background.
	Origen string `json:"origen" gorm:"type:varchar(20);not null"` // "manual" | "automatico"
	// Resultado es el desenlace del intento: "aceptado"/"rechazado" (el
	// facturador respondió) o "error" (fallo de transporte).
	Resultado string    `json:"resultado" gorm:"type:varchar(20);not null;index"`
	Mensaje   string    `json:"mensaje" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}

func (LogEnvio) TableName() string { return "logs_envio" }
