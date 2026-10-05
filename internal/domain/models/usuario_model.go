package models

import (
	"time"

	"gorm.io/gorm"
)

type Regional struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Nombre    string         `json:"nombre" gorm:"type:varchar(100);not null;uniqueIndex"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Regional) TableName() string { return "regionales" }

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

	Rol string `json:"rol" gorm:"type:varchar(20);not null;default:'operador'"`

	AccesoTotal bool `json:"acceso_total" gorm:"default:false"`

	SucursalesPermitidasCodigos string         `json:"sucursales_permitidas_codigos" gorm:"type:text"`
	CreatedAt                   time.Time      `json:"created_at"`
	UpdatedAt                   time.Time      `json:"updated_at"`
	DeletedAt                   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Usuario) TableName() string { return "usuarios" }

const (
	RolAdmin     = "admin"
	RolOperador  = "operador"
	RolConsultas = "consultas"
)
