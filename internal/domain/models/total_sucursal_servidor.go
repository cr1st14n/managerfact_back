package models

type SucursalTotal struct {
	Sucursal string  `json:"sucursal" gorm:"column:sucursal"`
	Facturas int     `json:"facturas" gorm:"column:facturas"`
	Monto    float64 `json:"monto" gorm:"column:monto"`
}

type ServidorTotales struct {
	Servidor   string          `json:"servidor"`
	Sucursales []SucursalTotal `json:"sucursales"`
	Error      string          `json:"error,omitempty"`
}
