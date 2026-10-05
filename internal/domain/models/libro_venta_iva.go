package models

type LibroVentaIva struct {
	Fecha              string  `json:"fecha" gorm:"column:fecha"`
	Sucursal           string  `json:"sucursal" gorm:"column:sucursal"`
	NumeroFactura      string  `json:"numero_factura" gorm:"column:numero_factura"`
	CUF                string  `json:"cuf" gorm:"column:cuf"`
	CodigoRecepcionSin string  `json:"codigo_recepcion_sin" gorm:"column:codigo_recepcion_sin"`
	NumeroDocumento    string  `json:"numero_documento" gorm:"column:numero_documento"`
	Complemento        string  `json:"complemento" gorm:"column:complemento"`
	NombreRazonSocial  string  `json:"nombre_razon_social" gorm:"column:nombre_razon_social"`
	MontoTotal         float64 `json:"monto_total" gorm:"column:monto_total"`
	Ice                float64 `json:"ice" gorm:"column:ice"`
	Descuentos         float64 `json:"descuentos" gorm:"column:descuentos"`
	BaseDebitoFiscal   float64 `json:"base_debito_fiscal" gorm:"column:base_debito_fiscal"`
	DebitoFiscal       float64 `json:"debito_fiscal" gorm:"column:debito_fiscal"`

	EstadoLibro           string `json:"estado_libro" gorm:"column:estado_libro"`
	EstadoDocumentoFiscal string `json:"estado_documento_fiscal" gorm:"column:estado_documento_fiscal"`
}

type LibroVentaIvaResumenSucursal struct {
	Sucursal         string  `json:"sucursal" gorm:"column:sucursal"`
	FacturasValidas  int     `json:"facturas_validas" gorm:"column:facturas_validas"`
	FacturasAnuladas int     `json:"facturas_anuladas" gorm:"column:facturas_anuladas"`
	MontoTotal       float64 `json:"monto_total" gorm:"column:monto_total"`
	BaseDebitoFiscal float64 `json:"base_debito_fiscal" gorm:"column:base_debito_fiscal"`
	DebitoFiscal     float64 `json:"debito_fiscal" gorm:"column:debito_fiscal"`
}
