package models

// IngresosPorMetodoPago es una fila de "Ingresos por Método de Pago".
// MetodoPago (el texto) puede venir vacío si no se pudo resolver contra la
// paramétrica SIN -- en ese caso el reporte igual sale con
// CodigoMetodoPago crudo. Ver ClicReportes.md sección 8.
type IngresosPorMetodoPago struct {
	CodigoMetodoPago string  `json:"codigo_metodo_pago" gorm:"column:codigo_metodo_pago"`
	MetodoPago       string  `json:"metodo_pago" gorm:"column:metodo_pago"`
	Facturas         int     `json:"facturas" gorm:"column:facturas"`
	TotalFacturado   float64 `json:"total_facturado" gorm:"column:total_facturado"`
}
