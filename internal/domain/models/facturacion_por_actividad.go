package models

// FacturacionPorActividad es una fila de "Facturación por Actividad
// Económica": total facturado y base del débito fiscal agrupados por la
// actividad económica de CABECERA de la factura (una por documento). Si
// NAABOL factura ítems de actividades distintas en una misma factura, esta
// vista asigna todo a una sola actividad -- para separar exacto habría que
// ir por detalle, pero sfe_detalle_documento_fiscal.actividad_economica es
// texto libre, por eso no se usa como llave. Ver ClicReportes.md sección 5.
type FacturacionPorActividad struct {
	CodigoActividadEconomica string  `json:"codigo_actividad_economica" gorm:"column:codigo_actividad_economica"`
	ActividadEconomica       string  `json:"actividad_economica" gorm:"column:actividad_economica"`
	Facturas                 int     `json:"facturas" gorm:"column:facturas"`
	TotalFacturado           float64 `json:"total_facturado" gorm:"column:total_facturado"`
	BaseDebitoFiscal         float64 `json:"base_debito_fiscal" gorm:"column:base_debito_fiscal"`
}
