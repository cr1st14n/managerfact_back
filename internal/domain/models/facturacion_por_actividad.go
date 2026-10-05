package models

type FacturacionPorActividad struct {
	CodigoActividadEconomica string  `json:"codigo_actividad_economica" gorm:"column:codigo_actividad_economica"`
	ActividadEconomica       string  `json:"actividad_economica" gorm:"column:actividad_economica"`
	Facturas                 int     `json:"facturas" gorm:"column:facturas"`
	TotalFacturado           float64 `json:"total_facturado" gorm:"column:total_facturado"`
	BaseDebitoFiscal         float64 `json:"base_debito_fiscal" gorm:"column:base_debito_fiscal"`
}
