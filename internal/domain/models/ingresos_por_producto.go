package models

// IngresosPorProducto es una fila de "Ingresos por Servicio/Producto": el
// reporte que mapea a cuentas de ingreso, agrupado por producto (no por
// factura). Ojo: SUM(subtotal) NO incluye el descuento adicional global de
// cabecera, así que el total puede ser mayor al de
// ResumenMensualDebitoFiscal -- la diferencia se explica en el reporte de
// cuadre cabecera vs. detalle (ClicReportes.md sección 15, pendiente). Ver
// ClicReportes.md sección 6.
type IngresosPorProducto struct {
	CodigoProductoSfe  string  `json:"codigo_producto_sfe" gorm:"column:codigo_producto_sfe"`
	CodigoProductoSin  string  `json:"codigo_producto_sin" gorm:"column:codigo_producto_sin"`
	DescripcionFactura string  `json:"descripcion_factura" gorm:"column:descripcion_factura"`
	DescripcionSin     string  `json:"descripcion_sin" gorm:"column:descripcion_sin"`
	Cantidad           float64 `json:"cantidad" gorm:"column:cantidad"`
	Subtotal           float64 `json:"subtotal" gorm:"column:subtotal"`
	DescuentoItem      float64 `json:"descuento_item" gorm:"column:descuento_item"`
}
