package models

// package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	// "gorm.io/gorm"
)

// DbConnection representa la configuración de conexión a bases de datos SQL Server
type DbConnection struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	ServerName   string         `json:"server_name" gorm:"type:varchar(100);not null;uniqueIndex" validate:"required,min=3,max=100"`
	Host         string         `json:"host" gorm:"type:varchar(255);not null" validate:"required,hostname_rfc1123"`
	Port         int            `json:"port" gorm:"not null;default:1433" validate:"required,min=1,max=65535"`
	DatabaseName string         `json:"database_name" gorm:"type:varchar(100);not null" validate:"required,min=1,max=100"`
	Username     string         `json:"username" gorm:"type:varchar(100);not null" validate:"required,min=1,max=100"`
	Password     string         `json:"password" gorm:"type:varchar(255);not null" validate:"required,min=1"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	Description  string         `json:"description" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName especifica el nombre de la tabla
func (DbConnection) TableName() string {
	return "db_connections"
}

// BeforeCreate se ejecuta antes de crear un registro
func (dc *DbConnection) BeforeCreate(tx *gorm.DB) error {
	// Aquí se podría encriptar la contraseña si es necesario
	return nil
}

// ConnectionString construye la cadena de conexión para SQL Server
func (dc *DbConnection) ConnectionString() string {
	return fmt.Sprintf(
		"server=%s;port=%d;database=%s;user id=%s;password=%s;encrypt=false;connection timeout=30",
		dc.Host, dc.Port, dc.DatabaseName, dc.Username, dc.Password,
	)
}

// IsValid verifica si la configuración tiene los campos requeridos
func (dc *DbConnection) IsValid() bool {
	return dc.ServerName != "" &&
		dc.Host != "" &&
		dc.Port > 0 &&
		dc.DatabaseName != "" &&
		dc.Username != "" &&
		dc.Password != ""
}
