package models

type IngresosPorMetodoPago struct {
	CodigoMetodoPago string  `json:"codigo_metodo_pago" gorm:"column:codigo_metodo_pago"`
	MetodoPago       string  `json:"metodo_pago" gorm:"column:metodo_pago"`
	Facturas         int     `json:"facturas" gorm:"column:facturas"`
	TotalFacturado   float64 `json:"total_facturado" gorm:"column:total_facturado"`
}
