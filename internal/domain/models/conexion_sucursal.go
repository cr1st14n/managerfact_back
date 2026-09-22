package models

import "time"

// ConexionSucursal es la copia local (en nuestro Postgres) de una fila de
// sfe_sucursal del SQL Server de una conexión "facturador". Existe para que
// Reportes y Descarga Mensual listen las sucursales sin conectarse a
// producción en cada selección; el admin la refresca a mano desde
// /conexiones. Solo guarda las columnas que usa el front.
//
// Los campos van como string (salvo SfeID) para conservar el mismo JSON que
// devolvía la consulta en vivo (models.SFE_sucursales).
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
