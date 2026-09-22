package models

// FacturaAnuladaSin es una fila de "Facturas Anuladas del Período": cada
// solicitud de anulación registrada en el SIN, filtrada por la fecha en que
// se envió la anulación (NO la fecha de emisión de la factura original) --
// una factura de un mes anulada al mes siguiente aparece en el mes de la
// anulación, por eso se traen ambas fechas. Puede haber varias filas por
// factura (reintentos): no se deduplica, para ver el historial completo.
// Ver ClicReportes.md sección 3.
//
// No confundir con models.FacturaAnulacion: ese es el pedido de anulación
// que arma esta app (Excel -> facturador, en Postgres); este es el registro
// que ya existe en el SQL Server del cliente, de solo lectura.
type FacturaAnuladaSin struct {
	NumeroFactura         string  `json:"numero_factura" gorm:"column:numero_factura"`
	CUF                   string  `json:"cuf" gorm:"column:cuf"`
	FechaEmision          string  `json:"fecha_emision" gorm:"column:fecha_emision"`
	FechaAnulacion        string  `json:"fecha_anulacion" gorm:"column:fecha_anulacion"`
	NumeroDocumento       string  `json:"numero_documento" gorm:"column:numero_documento"`
	NombreRazonSocial     string  `json:"nombre_razon_social" gorm:"column:nombre_razon_social"`
	MontoTotal            float64 `json:"monto_total" gorm:"column:monto_total"`
	CodigoMotivo          string  `json:"codigo_motivo" gorm:"column:codigo_motivo"`
	EstadoAnulacion       string  `json:"estado_anulacion" gorm:"column:estado_anulacion"`
	CodigoRespuestaSin    string  `json:"codigo_respuesta_sin" gorm:"column:codigo_respuesta_sin"`
	EstadoDocumentoFiscal string  `json:"estado_documento_fiscal" gorm:"column:estado_documento_fiscal"`
}
