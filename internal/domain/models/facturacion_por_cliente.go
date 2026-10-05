package models

type FacturacionPorCliente struct {
	NumeroDocumento   string  `json:"numero_documento" gorm:"column:numero_documento"`
	NombreRazonSocial string  `json:"nombre_razon_social" gorm:"column:nombre_razon_social"`
	Facturas          int     `json:"facturas" gorm:"column:facturas"`
	TotalFacturado    float64 `json:"total_facturado" gorm:"column:total_facturado"`
	PrimeraFactura    string  `json:"primera_factura" gorm:"column:primera_factura"`
	UltimaFactura     string  `json:"ultima_factura" gorm:"column:ultima_factura"`
}
