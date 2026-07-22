package models

import "time"

// DuasBusquedaParams representa los parámetros de búsqueda para DUAS
type DuasBusquedaParams struct {
	Nombre      string `json:"nombre"`
	Apellido    string `json:"apellido"`
	NumeroVuelo string `json:"numero_vuelo"`
	FechaDesde  string `json:"fecha_desde"`
	FechaHasta  string `json:"fecha_hasta"`
	Asiento     string `json:"asiento"`
	Ticket      string `json:"ticket"`
}

// DuasResultado representa un registro de la búsqueda DUAS
type DuasResultadoCentral struct {
	IDTESFacturaItinerario string    `json:"idtes_factura_itinerario" gorm:"column:IDTES_FACTURA_ITINERARIO"`
	FACNroFactura          string    `json:"fac_nrofactura" gorm:"column:NUMEROFACTURA"`
	FACNroVuelo            string    `json:"fac_nrovuelo" gorm:"column:FAC_NROVUELO"`
	FACFechaHoraVuelo      time.Time `json:"fac_fechahora_vuelo" gorm:"column:FAC_FECHAHORA_VUELO"`
	FACMonto               float64   `json:"fac_monto" gorm:"column:FAC_MONTO"`
	IDEstadoFactura        string    `json:"id_estadofactura" gorm:"column:IDA_ESTADOFACTURA"`
	FechaCreacion          time.Time `json:"fechacreacion" gorm:"column:FECHACREACION"`
	UsuarioCreacion        string    `json:"usuariocreacion" gorm:"column:USUARIOCREACION"`
	URLSin                 string    `json:"url_sin" gorm:"column:URL_SIN"`
	FACDetalleFactura      string    `json:"fac_detallefactura" gorm:"column:FAC_DETALLEFACTURA"`
	//FAC_FECHAEMISION_FACTURA
	FacFechaEmisionFactura time.Time `json:"fac_fechaemision_factura" gorm:"column:FAC_FECHAEMISION_FACTURA"`
}

// DuasResultadoLocal representa un registro de la búsqueda DUAS para facturas locales
type DuasResultadoLocal struct {
	IDTESFacturaItinerario string    `json:"idtes_factura_itinerario" gorm:"column:IDTES_FACTURA_ITINERARIO"`
	FACNroVuelo            string    `json:"fac_nrovuelo" gorm:"column:FAC_NROVUELO"`
	FacFechaEmisionFactura time.Time `json:"fac_fechaemision_factura" gorm:"column:FAC_FECHAEMISION_FACTURA"`
	FACDetalleFactura      string    `json:"fac_detallefactura" gorm:"column:FAC_DETALLEFACTURA"`
	FACMonto               float64   `json:"fac_monto" gorm:"column:FAC_MONTO"`
	FechaCreacion          time.Time `json:"fechacreacion" gorm:"column:FECHACREACION"`
	IDTESFacturaOnline     string    `json:"idtes_facturaonline" gorm:"column:IDTES_FACTURAONLINE"`
	Codigo                 string    `json:"codigo" gorm:"column:CODIGO"`
	IDDocumento            string    `json:"id_documento" gorm:"column:ID_DOCUMENTO"`
	CUF                    string    `json:"cuf" gorm:"column:CUF"`
	CUFD                   string    `json:"cufd" gorm:"column:CUFD"`
	CUIS                   string    `json:"cuis" gorm:"column:CUIS"`
	NumeroFactura          string    `json:"numerofactura" gorm:"column:NUMEROFACTURA"`
	FechaEmision           time.Time `json:"fecha_emision" gorm:"column:FECHA_EMISION"`
	EstadoDocumentoFiscal  string    `json:"estado_documento_fiscal" gorm:"column:ESTADO_DOCUMENTO_FISCAL"`
	CodigoIntegracion      string    `json:"codigo_integracion" gorm:"column:CODIGO_INTEGRACION"`
	URLSin                 string    `json:"url_sin" gorm:"column:URL_SIN"`
}

type DuasResultado struct {
	ResultadosLocal   []DuasResultadoLocal
	ResultadosCentral []DuasResultadoCentral
}
