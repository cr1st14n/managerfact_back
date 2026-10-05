package models

type IngresosPorProducto struct {
	CodigoProductoSfe  string  `json:"codigo_producto_sfe" gorm:"column:codigo_producto_sfe"`
	CodigoProductoSin  string  `json:"codigo_producto_sin" gorm:"column:codigo_producto_sin"`
	DescripcionFactura string  `json:"descripcion_factura" gorm:"column:descripcion_factura"`
	DescripcionSin     string  `json:"descripcion_sin" gorm:"column:descripcion_sin"`
	Cantidad           float64 `json:"cantidad" gorm:"column:cantidad"`
	Subtotal           float64 `json:"subtotal" gorm:"column:subtotal"`
	DescuentoItem      float64 `json:"descuento_item" gorm:"column:descuento_item"`
}
