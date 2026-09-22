package models

// IngresosPorMoneda es una fila de "Ingresos por Moneda y Tipo de Cambio".
// En bolivianos, TipoCambio = 1 y TotalEnMonedaOriginal = TotalBs. Para
// otras monedas, TotalEnMonedaOriginal está en moneda original y TotalBs en
// Bs. DiferenciaCambioAbs es valor absoluto (no trae signo): no asumir si
// es ganancia o pérdida. Ver ClicReportes.md sección 9.
type IngresosPorMoneda struct {
	CodigoMoneda          string  `json:"codigo_moneda" gorm:"column:codigo_moneda"`
	Moneda                string  `json:"moneda" gorm:"column:moneda"`
	TipoCambio            float64 `json:"tipo_cambio" gorm:"column:tipo_cambio"`
	TipoCambioOficial     float64 `json:"tipo_cambio_oficial" gorm:"column:tipo_cambio_oficial"`
	Facturas              int     `json:"facturas" gorm:"column:facturas"`
	TotalEnMonedaOriginal float64 `json:"total_en_moneda_original" gorm:"column:total_en_moneda_original"`
	TotalBs               float64 `json:"total_bs" gorm:"column:total_bs"`
	DiferenciaCambioAbs   float64 `json:"diferencia_cambio_abs" gorm:"column:diferencia_cambio_abs"`
}
