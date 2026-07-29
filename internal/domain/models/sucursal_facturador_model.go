package models

import (
	"time"

	"gorm.io/gorm"
)

// SucursalFacturador es la configuración de conexión al facturador
// (FacturaClic) de una sucursal. Es independiente del catálogo maestro
// SucursalCatalogo (que solo maneja accesos de usuario) — ver
// doc/EnvioFacturacion.md sección 1.
type SucursalFacturador struct {
	ID                uint   `json:"id" gorm:"primaryKey"`
	Nombre            string `json:"nombre" gorm:"type:varchar(150);not null"`
	CodigoSucursalSin int    `json:"codigo_sucursal_sin" gorm:"not null"`
	PuntoVentaEmisor  string `json:"punto_venta_emisor" gorm:"type:varchar(20)"`
	UrlLinkFacturador string `json:"url_link_facturador" gorm:"type:varchar(255);not null"`
	// TokenAcceso se guarda cifrado (AES-256-GCM) y nunca se serializa
	// directamente en JSON: el handler expone únicamente un booleano
	// "token_configurado" en la respuesta.
	TokenAcceso     string `json:"-" gorm:"type:varchar(500)"`
	CodigoMonedaBob string `json:"codigo_moneda_bob" gorm:"type:varchar(10)"`
	CodigoCI        string `json:"codigo_ci" gorm:"type:varchar(20)"`
	CodigoNit       string `json:"codigo_nit" gorm:"type:varchar(20);not null"`
	Activo          bool   `json:"activo" gorm:"default:true"`
	// EstadoConexion es el circuit breaker del EnvioWorker: "activo" (por
	// defecto) o "en_revision" cuando el último intento de envío falló por
	// un error de transporte (facturador caído/inalcanzable, no un rechazo
	// de negocio). Vuelve solo a "activo" en cuanto el facturador vuelve a
	// responder — ver doc/EnvioFacturacion.md sección 5.
	EstadoConexion      string         `json:"estado_conexion" gorm:"type:varchar(20);not null;default:'activo'"`
	UltimoErrorConexion *time.Time     `json:"ultimo_error_conexion"`
	UltimoErrorMensaje  string         `json:"ultimo_error_mensaje" gorm:"type:text"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

func (SucursalFacturador) TableName() string { return "sucursales_facturador" }
