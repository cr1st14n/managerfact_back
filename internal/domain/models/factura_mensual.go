package models

import "time"

// FacturaMensual es una fila de la descarga mensual por código de producto.
type FacturaMensual struct {
	NumeroFactura     string    `json:"numero_factura" gorm:"column:numero_factura"`
	FechaEmision      time.Time `json:"fecha_emision" gorm:"column:fecha_emision"`
	CUF               string    `json:"cuf" gorm:"column:cuf"`
	CUFD              string    `json:"cufd" gorm:"column:cufd"`
	MontoTotal        float64   `json:"monto_total" gorm:"column:monto_total"`
	CodigoProductoSfe string    `json:"codigo_producto_sfe" gorm:"column:codigo_producto_sfe"`
	Descripcion       string    `json:"descripcion" gorm:"column:descripcion"`
}
