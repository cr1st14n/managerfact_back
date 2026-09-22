package models

// EmisionPorUsuario es una fila de "Emisión por Usuario" (arqueo / control
// interno): total por usuario_emision + sucursal + día. USUARIO_EMISION es
// texto libre que manda el sistema integrador, no siempre es un usuario
// real (puede ser una cuenta de servicio de otro sistema) -- tomarlo como
// pista, no como evidencia. Ver ClicReportes.md sección 11.
type EmisionPorUsuario struct {
	UsuarioEmision string  `json:"usuario_emision" gorm:"column:usuario_emision"`
	Sucursal       string  `json:"sucursal" gorm:"column:sucursal"`
	Dia            string  `json:"dia" gorm:"column:dia"`
	Facturas       int     `json:"facturas" gorm:"column:facturas"`
	TotalFacturado float64 `json:"total_facturado" gorm:"column:total_facturado"`
}
