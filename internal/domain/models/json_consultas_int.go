package models

type Json_consulta_data struct {
	IdFacturador    string `json:"idServer"`
	NumeroDocumento string `json:"nit"`
	NumeroFactura   string `json:"numeroFactura"`
	CodigoProducto  string `json:"codigoProducto"`
	FechaDesde      string `json:"fechaDesde"`
	FechaHasta      string `json:"fechaHasta"`
	// Nit             string `json:"nit"`
	Sucursal string `json:"sucursal"`
}
