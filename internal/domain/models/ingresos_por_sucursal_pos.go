package models

// IngresosPorSucursalPos es una fila de "Ingresos por Sucursal y Punto de
// Venta": separa por aeropuerto/caja. CodigoPos y PuntoVenta vienen vacíos
// cuando la factura no tiene punto de venta asociado (LEFT JOIN a
// propósito: con INNER esas facturas se perderían del total). Ver
// ClicReportes.md sección 7.
type IngresosPorSucursalPos struct {
	CodigoSucursal        string  `json:"codigo_sucursal" gorm:"column:codigo_sucursal"`
	Sucursal              string  `json:"sucursal" gorm:"column:sucursal"`
	MunicipioDepartamento string  `json:"municipio_departamento" gorm:"column:municipio_departamento"`
	CodigoPos             string  `json:"codigo_pos" gorm:"column:codigo_pos"`
	PuntoVenta            string  `json:"punto_venta" gorm:"column:punto_venta"`
	Facturas              int     `json:"facturas" gorm:"column:facturas"`
	TotalFacturado        float64 `json:"total_facturado" gorm:"column:total_facturado"`
	BaseDebitoFiscal      float64 `json:"base_debito_fiscal" gorm:"column:base_debito_fiscal"`
}
