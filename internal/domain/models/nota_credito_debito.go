package models

type NotaCreditoDebito struct {
	Fecha                 string  `json:"fecha" gorm:"column:fecha"`
	NroNota               string  `json:"nro_nota" gorm:"column:nro_nota"`
	TipoDocumentoSector   int     `json:"tipo_documento_sector" gorm:"column:tipo_documento_sector"`
	NumeroFacturaOriginal string  `json:"numero_factura_original" gorm:"column:numero_factura_original"`
	CufFacturaOriginal    string  `json:"cuf_factura_original" gorm:"column:cuf_factura_original"`
	FechaFacturaOriginal  string  `json:"fecha_factura_original" gorm:"column:fecha_factura_original"`
	NumeroDocumento       string  `json:"numero_documento" gorm:"column:numero_documento"`
	NombreRazonSocial     string  `json:"nombre_razon_social" gorm:"column:nombre_razon_social"`
	MontoTotalOriginal    float64 `json:"monto_total_original" gorm:"column:monto_total_original"`
	MontoTotalDevuelto    float64 `json:"monto_total_devuelto" gorm:"column:monto_total_devuelto"`
	MontoTotalConciliado  float64 `json:"monto_total_conciliado" gorm:"column:monto_total_conciliado"`
	DebitoFiscalIva       float64 `json:"debito_fiscal_iva" gorm:"column:debito_fiscal_iva"`
	CreditoFiscalIva      float64 `json:"credito_fiscal_iva" gorm:"column:credito_fiscal_iva"`
	EstadoDocumentoFiscal string  `json:"estado_documento_fiscal" gorm:"column:estado_documento_fiscal"`
}
