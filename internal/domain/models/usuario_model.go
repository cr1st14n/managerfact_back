package models

import (
	"time"

	"gorm.io/gorm"
)

// Regional agrupa sucursales por zona geográfica (La Paz, Cochabamba, Santa
// Cruz, Beni). Ver doc/sucursales.md para el detalle de la agrupación.
type Regional struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Nombre    string         `json:"nombre" gorm:"type:varchar(100);not null;uniqueIndex"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Regional) TableName() string { return "regionales" }

// SucursalCatalogo es el catálogo maestro de sucursales/aeropuertos por
// código SIN. Es independiente de cada conexión de facturador: el
// codigo_sucursal_sin es único a nivel nacional y se repite igual en el
// sfe_sucursal de cada servidor.
type SucursalCatalogo struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	CodigoSucursalSin int            `json:"codigo_sucursal_sin" gorm:"not null;uniqueIndex"`
	Nombre            string         `json:"nombre" gorm:"type:varchar(150);not null"`
	RegionalID        uint           `json:"regional_id" gorm:"not null"`
	Regional          Regional       `json:"regional,omitempty" gorm:"foreignKey:RegionalID"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

func (SucursalCatalogo) TableName() string { return "sucursales_catalogo" }

// Usuario representa a un operador del sistema, autenticado con
// codigo_usuario + password (login vía JWT, ver pkg/utils/jwt.go).
type Usuario struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	Nombre        string            `json:"nombre" gorm:"type:varchar(150);not null"`
	CI            string            `json:"ci" gorm:"type:varchar(20);not null;uniqueIndex"`
	Cargo         string            `json:"cargo" gorm:"type:varchar(100)"`
	CodigoUsuario string            `json:"codigo_usuario" gorm:"type:varchar(50);not null;uniqueIndex"`
	PasswordHash  string            `json:"-" gorm:"type:varchar(255);not null"`
	RegionalID    *uint             `json:"regional_id"`
	Regional      *Regional         `json:"regional,omitempty" gorm:"foreignKey:RegionalID"`
	SucursalID    *uint             `json:"sucursal_id"`
	Sucursal      *SucursalCatalogo `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
	IsActive      bool              `json:"is_active" gorm:"default:true"`
	// AccesoTotal otorga acceso a todas las sucursales sin importar
	// SucursalesPermitidasCodigos.
	AccesoTotal bool `json:"acceso_total" gorm:"default:false"`
	// SucursalesPermitidasCodigos es la colección de códigos SIN de sucursal
	// a los que puede acceder este usuario, como texto separado por comas
	// (ej. "1,6,7") — reemplaza las antiguas tablas de acceso por
	// regional/sucursal (usuario_accesos_regional/usuario_accesos_sucursal).
	// Verificar acceso a una sucursal es una simple búsqueda del código SIN
	// dentro de este texto (ver UsuarioService.TieneAccesoSucursal). El
	// front sigue armando la selección por regional/sucursal vía IDs (más
	// cómodo de mostrar agrupado); el backend resuelve esos IDs a códigos
	// SIN al guardar.
	SucursalesPermitidasCodigos string         `json:"sucursales_permitidas_codigos" gorm:"type:text"`
	CreatedAt                   time.Time      `json:"created_at"`
	UpdatedAt                   time.Time      `json:"updated_at"`
	DeletedAt                   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Usuario) TableName() string { return "usuarios" }
