package models

import (
	"time"

	"gorm.io/gorm"
)

// FacturaPrevalorada es un boleto importado (etapa 1) que después se envía
// al facturador de su sucursal (etapa 2) — ver doc/EnvioFacturacion.md
// sección 2. Siempre trae un solo ítem (derecho aeroportuario), por eso se
// modela plano en una sola tabla, sin tabla hija de detalle.
type FacturaPrevalorada struct {
	ID                   uint                `json:"id" gorm:"primaryKey"`
	SucursalFacturadorID uint                `json:"sucursal_facturador_id" gorm:"not null"`
	SucursalFacturador   *SucursalFacturador `json:"sucursal_facturador,omitempty" gorm:"foreignKey:SucursalFacturadorID"`
	LoteID               string              `json:"lote_id" gorm:"type:varchar(36);not null;index"`
	CodigoIntegracion    string              `json:"codigo_integracion" gorm:"type:varchar(64);not null;uniqueIndex"`
	// Tipo es fijo "FACTURA_PREVALORADA": no viene del Excel, existe para
	// acomodar otros tipos de factura a futuro sin construir hoy una
	// abstracción genérica.
	Tipo string `json:"tipo" gorm:"type:varchar(30);not null;default:'FACTURA_PREVALORADA'"`
	// Observacion es el motivo de carga del lote (por qué se importó este
	// archivo), ingresado una sola vez antes de importar y fijo para todas
	// las filas del lote — mismo patrón que SucursalFacturadorID.
	Observacion string `json:"observacion" gorm:"type:varchar(255);not null"`

	// Etapa 1: datos importados del Excel/lista de boletos.
	Detalle           string    `json:"detalle" gorm:"type:varchar(255);not null"`
	CodigoProducto    string    `json:"codigo_producto" gorm:"type:varchar(30);not null"`
	CostoDuaDolares   float64   `json:"costo_dua_dolares" gorm:"not null"`
	FechaCompraBoleto time.Time `json:"fecha_compra_boleto" gorm:"not null"`
	// TipoCambio es el tc vigente a la fecha de FechaCompraBoleto, tal cual
	// viene del Excel — no se recalcula.
	TipoCambio   float64   `json:"tipo_cambio" gorm:"not null"`
	FechaEmision time.Time `json:"fecha_emision" gorm:"not null"`

	// Etapa 2: seguimiento de envío al facturador.
	Estado           string     `json:"estado" gorm:"type:varchar(20);not null;default:'pendiente';index"`
	CodigoRespuesta  string     `json:"codigo_respuesta" gorm:"type:varchar(50)"`
	MensajeRespuesta string     `json:"mensaje_respuesta" gorm:"type:text"`
	FechaEnvio       *time.Time `json:"fecha_envio"`
	FechaRespuesta   *time.Time `json:"fecha_respuesta"`
	IntentosConsulta int        `json:"intentos_consulta" gorm:"default:0"`
	// CUF identifica el documento fiscal ya aceptado por el SIN — se
	// necesita más adelante para armar el Excel de anulación (sección 4).
	CUF           string `json:"cuf" gorm:"type:varchar(100)"`
	NumeroFactura string `json:"numero_factura" gorm:"type:varchar(50)"`
	UrlDocumento  string `json:"url_documento" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (FacturaPrevalorada) TableName() string { return "facturas_prevaloradas" }
