package models

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
