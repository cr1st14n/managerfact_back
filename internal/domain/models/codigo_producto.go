package models

import (
	"time"

	"gorm.io/gorm"
)

type Codigo_producto struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Codigo      string         `json:"codigo"`
	Descripcion string         `json:"descripcion"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (Codigo_producto) TableName() string {
	return "db_codigo_producto"
}
