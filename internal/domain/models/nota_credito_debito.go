package models

// NotaCreditoDebito es una fila de "Notas de Crédito/Débito y
// Conciliación": documentos del sector 24 (nota crédito-débito) o 29 (nota
// de conciliación) según el catálogo SIN, más cualquier documento que
// apunte a una factura original como respaldo (por si el sector real
// difiere de la asunción). DebitoFiscalIva/CreditoFiscalIva ya vienen
// calculados en el documento: no recalcular el 13% acá. Ver
// ClicReportes.md sección 4 -- incluye la advertencia "ASUNCION A VALIDAR"
// sobre los códigos de sector 24/29, confirmar contra la sección 0.2 antes
// de confiar en el desglose por sector.
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
