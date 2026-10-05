package models

import "time"

type ConexionSucursal struct {
	ID         uint  `gorm:"primaryKey"`
	ConexionID uint  `gorm:"not null;uniqueIndex:idx_conexion_sucursal"`
	SfeID      int64 `gorm:"column:sfe_id;not null;uniqueIndex:idx_conexion_sucursal"`

	CodigoSucursal        string
	CodigoSucursalSin     string
	Nombre                string
	Direccion             string
	MunicipioDepartamento string
	EstadoSucursal        string

	ActualizadoAt time.Time
}

func (ConexionSucursal) TableName() string {
	return "conexion_sucursales"
}
