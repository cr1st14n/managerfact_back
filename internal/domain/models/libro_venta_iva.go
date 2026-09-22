package models

// LibroVentaIva es una fila del Libro de Ventas IVA: detalle factura por
// factura de TODAS las sucursales de una conexión (empresa) en un período.
// Replica la consulta manual de ClicReportes.md sección 1 -- ver ahí el
// porqué de cada columna (la base del débito fiscal es
// monto_total_sujeto_iva, no monto_total; las anuladas se dejan con el
// monto original en vez de en cero, para poder cuadrar). Los nombres de
// columna vienen del diccionario v0.2 de NAABOL: si alguno falla al correr
// contra un servidor nuevo, revisar con sp_help 'sfe_documento_fiscal'
// antes de asumir que el dato no existe.
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
	// "V" verificada / "A" anulada -- a diferencia del libro oficial (que
	// pone las anuladas en cero), aquí se deja el monto original para poder
	// cuadrar contra el sistema.
	EstadoLibro           string `json:"estado_libro" gorm:"column:estado_libro"`
	EstadoDocumentoFiscal string `json:"estado_documento_fiscal" gorm:"column:estado_documento_fiscal"`
}

// LibroVentaIvaResumenSucursal es el total de control por sucursal del
// Libro de Ventas IVA: mismo débito fiscal que el detalle, pero agregado en
// SQL (GROUP BY) para no tener que traer factura por factura solo para
// sumarlas -- la facturación de una empresa puede ser miles de filas al
// mes. El detalle completo (LibroVentaIva) queda solo para la descarga a
// Excel, nunca para pintar en pantalla.
type LibroVentaIvaResumenSucursal struct {
	Sucursal         string  `json:"sucursal" gorm:"column:sucursal"`
	FacturasValidas  int     `json:"facturas_validas" gorm:"column:facturas_validas"`
	FacturasAnuladas int     `json:"facturas_anuladas" gorm:"column:facturas_anuladas"`
	MontoTotal       float64 `json:"monto_total" gorm:"column:monto_total"`
	BaseDebitoFiscal float64 `json:"base_debito_fiscal" gorm:"column:base_debito_fiscal"`
	DebitoFiscal     float64 `json:"debito_fiscal" gorm:"column:debito_fiscal"`
}
