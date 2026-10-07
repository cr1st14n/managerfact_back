package models

import (
	"time"

	"gorm.io/gorm"
)

type FacturaPrevalorada struct {
	ID                   uint                `json:"id" gorm:"primaryKey"`
	SucursalFacturadorID uint                `json:"sucursal_facturador_id" gorm:"not null"`
	SucursalFacturador   *SucursalFacturador `json:"sucursal_facturador,omitempty" gorm:"foreignKey:SucursalFacturadorID"`
	LoteID               string              `json:"lote_id" gorm:"type:varchar(36);not null;index"`
	CodigoIntegracion    string              `json:"codigo_integracion" gorm:"type:varchar(64);not null;index"`

	Tipo string `json:"tipo" gorm:"type:varchar(30);not null;default:'FACTURA_PREVALORADA'"`

	Observacion string `json:"observacion" gorm:"type:varchar(255);not null"`

	Detalle         string  `json:"detalle" gorm:"type:varchar(255);not null"`
	CodigoProducto  string  `json:"codigo_producto" gorm:"type:varchar(30);not null"`
	CostoDuaDolares float64 `json:"costo_dua_dolares" gorm:"not null"`
		// El tipo date evita que UTC-4 desplace un día las fechas sin hora importadas del Excel.
	FechaCompraBoleto time.Time `json:"fecha_compra_boleto" gorm:"type:date;not null"`

	TipoCambio float64 `json:"tipo_cambio" gorm:"not null"`
	TotalBob     float64   `json:"total_bob" gorm:"not null;default:0"`
	FechaEmision time.Time `json:"fecha_emision" gorm:"type:date;not null"`

	Estado           string     `json:"estado" gorm:"type:varchar(20);not null;default:'pendiente';index"`
	CodigoRespuesta  string     `json:"codigo_respuesta" gorm:"type:varchar(50)"`
	MensajeRespuesta string     `json:"mensaje_respuesta" gorm:"type:text"`
	FechaEnvio       *time.Time `json:"fecha_envio"`
	FechaRespuesta   *time.Time `json:"fecha_respuesta"`
	IntentosConsulta int        `json:"intentos_consulta" gorm:"default:0"`

	CUF           string `json:"cuf" gorm:"type:varchar(100)"`
	NumeroFactura string `json:"numero_factura" gorm:"type:varchar(50)"`
	UrlDocumento  string `json:"url_documento" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (FacturaPrevalorada) TableName() string { return "facturas_prevaloradas" }
