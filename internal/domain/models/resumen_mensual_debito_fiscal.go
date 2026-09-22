package models

// ResumenMensualDebitoFiscal es el "total de control" mensual del Libro de
// Ventas IVA: una fila por año/mes con el mismo débito fiscal que el libro,
// pero de TODA la empresa (sin desglosar por sucursal, a diferencia de
// LibroVentaIvaResumenSucursal). Debe coincidir con la suma de
// LibroVentaIva y con lo declarado al SIN; si no cuadra, casi siempre es un
// estado no contemplado o facturas de un mes emitidas en el siguiente
// (fecha de emisión vs fecha de envío). Ver ClicReportes.md sección 2.
type ResumenMensualDebitoFiscal struct {
	Anio             int     `json:"anio" gorm:"column:anio"`
	Mes              int     `json:"mes" gorm:"column:mes"`
	FacturasValidas  int     `json:"facturas_validas" gorm:"column:facturas_validas"`
	FacturasAnuladas int     `json:"facturas_anuladas" gorm:"column:facturas_anuladas"`
	TotalFacturado   float64 `json:"total_facturado" gorm:"column:total_facturado"`
	BaseDebitoFiscal float64 `json:"base_debito_fiscal" gorm:"column:base_debito_fiscal"`
	DebitoFiscal     float64 `json:"debito_fiscal" gorm:"column:debito_fiscal"`
}
