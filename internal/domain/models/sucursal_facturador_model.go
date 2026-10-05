package models

import (
	"time"

	"gorm.io/gorm"
)

type SucursalFacturador struct {
	ID                uint   `json:"id" gorm:"primaryKey"`
	Nombre            string `json:"nombre" gorm:"type:varchar(150);not null"`
	CodigoSucursalSin int    `json:"codigo_sucursal_sin" gorm:"not null"`
	PuntoVentaEmisor  string `json:"punto_venta_emisor" gorm:"type:varchar(20)"`
	UrlLinkFacturador string `json:"url_link_facturador" gorm:"type:varchar(255);not null"`

	TokenAcceso     string `json:"-" gorm:"type:varchar(500)"`
	CodigoMonedaBob string `json:"codigo_moneda_bob" gorm:"type:varchar(10)"`
	CodigoCI        string `json:"codigo_ci" gorm:"type:varchar(20)"`
	CodigoNit       string `json:"codigo_nit" gorm:"type:varchar(20);not null"`
	Activo          bool   `json:"activo" gorm:"default:true"`
		// Solo errores de transporte activan en_revision; una respuesta de negocio reactiva la sucursal.
	EstadoConexion      string         `json:"estado_conexion" gorm:"type:varchar(20);not null;default:'activo'"`
	UltimoErrorConexion *time.Time     `json:"ultimo_error_conexion"`
	UltimoErrorMensaje  string         `json:"ultimo_error_mensaje" gorm:"type:text"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

func (SucursalFacturador) TableName() string { return "sucursales_facturador" }
