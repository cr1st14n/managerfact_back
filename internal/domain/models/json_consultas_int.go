package models

type Json_consulta_data struct {
	IdFacturador    string   `json:"idServer"`
	NumeroDocumento string   `json:"nit"`
	NumeroFactura   string   `json:"numeroFactura"`
	CodigoProducto  []string `json:"codigoProducto"`
	FechaDesde      string   `json:"fechaDesde"`
	FechaHasta      string   `json:"fechaHasta"`
	// Nit             string `json:"nit"`
	Sucursal string `json:"sucursal"`

	// Filtros adicionales del reporte completo de facturadores
	CodigoIntegracion     string `json:"codigoIntegracion"`
	CodigoCliente         string `json:"codigoCliente"`
	CUF                   string `json:"cuf"`
	EstadoDocumentoFiscal string `json:"estadoDocumentoFiscal"`
	CodigoSucursalSin     string `json:"codigoSucursalSin"`
	TipoEmision           string `json:"tipoEmision"`
}
