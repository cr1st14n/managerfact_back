package models

import "time"

// DuasBusquedaParams representa los parámetros de búsqueda para DUAS
type DuasBusquedaParams struct {
	Nombre     string `json:"nombre"`
	Apellido   string `json:"apellido"`
	NumeroVuelo string `json:"numero_vuelo"`
	FechaDesde string `json:"fecha_desde"`
	FechaHasta string `json:"fecha_hasta"`
	Asiento    string `json:"asiento"`
	Ticket     string `json:"ticket"`
}

// DuasResultado representa un registro de la búsqueda DUAS
type DuasResultado struct {
	IDTESFacturaItinerario string    `json:"idtes_factura_itinerario" gorm:"column:IDTES_FACTURA_ITINERARIO"`
	FACNroFactura          string    `json:"fac_nrofactura" gorm:"column:FAC_NROFACTURA"`
	FACNroVuelo            string    `json:"fac_nrovuelo" gorm:"column:FAC_NROVUELO"`
	FACFechaHoraVuelo      time.Time `json:"fac_fechahora_vuelo" gorm:"column:FAC_FECHAHORA_VUELO"`
	FACMonto               float64   `json:"fac_monto" gorm:"column:FAC_MONTO"`
	IDEstadoFactura        string    `json:"id_estadofactura" gorm:"column:IDA_ESTADOFACTURA"`
	FechaCreacion          time.Time `json:"fechacreacion" gorm:"column:FECHACREACION"`
	UsuarioCreacion        string    `json:"usuariocreacion" gorm:"column:USUARIOCREACION"`
	URLSin                 string    `json:"url_sin" gorm:"column:URL_SIN"`
	FACDetalleFactura      string    `json:"fac_detallefactura" gorm:"column:FAC_DETALLEFACTURA"`
	BCBPApellido           string    `json:"bcbp_apellido" gorm:"column:BCBP_APELLIDO"`
	BCBPNombre             string    `json:"bcbp_nombre" gorm:"column:BCBP_NOMBRE"`
	BCBPOrigen             string    `json:"bcbp_origen" gorm:"column:BCBP_ORIGEN"`
	BCBPDestino            string    `json:"bcbp_destino" gorm:"column:BCBP_DESTINO"`
	BCBPAerolinea          string    `json:"bcbp_aerolinea" gorm:"column:BCBP_AEROLINEA"`
	BCBPVuelo              string    `json:"bcbp_vuelo" gorm:"column:BCBP_VUELO"`
	BCBPDiaJuliano         string    `json:"bcbp_dia_juliano" gorm:"column:BCBP_DIA_JULIANO"`
	BCBPClase              string    `json:"bcbp_clase" gorm:"column:BCBP_CLASE"`
	BCBPAsiento            string    `json:"bcbp_asiento" gorm:"column:BCBP_ASIENTO"`
	BCBPsecuencia          string    `json:"bcbp_secuencia" gorm:"column:BCBP_SECUENCIA"`
	BCBPNroTicket          string    `json:"bcbp_nro_ticket" gorm:"column:BCBP_NRO_TICKET"`
}
