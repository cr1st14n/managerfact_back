package models

// FacturacionPorCliente es una fila de "Facturación por Cliente": agrupa
// por número de documento + razón social (no por id de cliente, porque el
// mismo cliente se puede repetir con distintos ids según cómo se integró).
// Consumidor final (documento 99002) sale como una fila más, no se excluye,
// para que el total cuadre con las otras secciones. Es lo FACTURADO, no lo
// cobrado. Ver ClicReportes.md sección 10.
//
// El número de clientes distintos de una empresa puede ser grande (mucha
// facturación): por eso el reporte siempre trae como mucho "top" filas,
// ordenadas por total facturado -- ver ResumenContableService.FacturacionPorCliente.
type FacturacionPorCliente struct {
	NumeroDocumento   string  `json:"numero_documento" gorm:"column:numero_documento"`
	NombreRazonSocial string  `json:"nombre_razon_social" gorm:"column:nombre_razon_social"`
	Facturas          int     `json:"facturas" gorm:"column:facturas"`
	TotalFacturado    float64 `json:"total_facturado" gorm:"column:total_facturado"`
	PrimeraFactura    string  `json:"primera_factura" gorm:"column:primera_factura"`
	UltimaFactura     string  `json:"ultima_factura" gorm:"column:ultima_factura"`
}
