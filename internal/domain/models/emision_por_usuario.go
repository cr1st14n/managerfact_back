package models

type EmisionPorUsuario struct {
	UsuarioEmision string  `json:"usuario_emision" gorm:"column:usuario_emision"`
	Sucursal       string  `json:"sucursal" gorm:"column:sucursal"`
	Dia            string  `json:"dia" gorm:"column:dia"`
	Facturas       int     `json:"facturas" gorm:"column:facturas"`
	TotalFacturado float64 `json:"total_facturado" gorm:"column:total_facturado"`
}
