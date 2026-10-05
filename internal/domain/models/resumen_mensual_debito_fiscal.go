package models

// Este total debe coincidir con el libro y lo declarado al SIN; revisar fechas de emisión/envío si difiere.
type ResumenMensualDebitoFiscal struct {
	Anio             int     `json:"anio" gorm:"column:anio"`
	Mes              int     `json:"mes" gorm:"column:mes"`
	FacturasValidas  int     `json:"facturas_validas" gorm:"column:facturas_validas"`
	FacturasAnuladas int     `json:"facturas_anuladas" gorm:"column:facturas_anuladas"`
	TotalFacturado   float64 `json:"total_facturado" gorm:"column:total_facturado"`
	BaseDebitoFiscal float64 `json:"base_debito_fiscal" gorm:"column:base_debito_fiscal"`
	DebitoFiscal     float64 `json:"debito_fiscal" gorm:"column:debito_fiscal"`
}
