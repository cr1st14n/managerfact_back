package models

// Los meses agrupan por la fecha elegida (tipoFecha), no siempre por emisión.
type ResumenMensualDebitoFiscal struct {
	Anio             int     `json:"anio" gorm:"column:anio"`
	Mes              int     `json:"mes" gorm:"column:mes"`
	FacturasValidas  int     `json:"facturas_validas" gorm:"column:facturas_validas"`
	FacturasAnuladas int     `json:"facturas_anuladas" gorm:"column:facturas_anuladas"`
	TotalFacturado   float64 `json:"total_facturado" gorm:"column:total_facturado"`
	BaseDebitoFiscal float64 `json:"base_debito_fiscal" gorm:"column:base_debito_fiscal"`
	DebitoFiscal     float64 `json:"debito_fiscal" gorm:"column:debito_fiscal"`
}
