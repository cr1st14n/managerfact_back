package models

import "time"

type FacturacionNaabol struct {
	ID                     uint      `gorm:"primaryKey;column:id"`
	CreatedBy              string    `gorm:"column:created_by"`
	CreatedDate            time.Time `gorm:"column:created_date"`
	Estado                 string    `gorm:"column:estado"`
	CodigoIntegracion      string    `gorm:"column:codigo_integracion"`
	CUF                    string    `gorm:"column:cuf"`
	CUFD                   string    `gorm:"column:cufd"`
	CUIS                   string    `gorm:"column:cuis"`
	EstadoDocumentoFiscal  string    `gorm:"column:estado_documento_fiscal"`
	FechaEmision           time.Time `gorm:"column:fecha_emision"`
	FechaEmisionFactura    time.Time `gorm:"column:fecha_emision_factura"`
	FechaEnvio             time.Time `gorm:"column:fecha_envio" json:"fecha_envio"`
	FechaRespuesta         time.Time `gorm:"column:fecha_respuesta"`
	Gestion                int       `gorm:"column:gestion"`
	IDPaquete              string    `gorm:"column:id_paquete"`
	IdentificadorURL       string    `gorm:"column:identificador_url"`
	Mes                    int       `gorm:"column:mes"`
	NombreRazonSocial      string    `gorm:"column:nombre_razon_social" json:"nombre_razon_social"`
	NumeroAutorizacionCUF  string    `gorm:"column:numero_autorizacion_cuf"`
	NumeroDocumento        string    `gorm:"column:numero_documento"`
	NumeroFactura          string    `gorm:"column:numero_factura"`
	NumeroFacturaOriginal  string    `gorm:"column:numero_factura_original"`
	Regenerate             string    `gorm:"column:regenerate"`
	TipoCambio             float64   `gorm:"column:tipo_cambio"`
	TipoEmision            string    `gorm:"column:tipo_emision"`
	TipoFactura            string    `gorm:"column:tipo_factura"`
	UsuarioEmision         string    `gorm:"column:usuario_emision"`
	ID_SFECliente          int       `gorm:"column:id_sfe_cliente"`
	MetodoPago             string    `gorm:"column:metodo_pago"`
	TipoDocumentoFiscal    string    `gorm:"column:tipo_documento_fiscal"`
	TipoDocumentoIdentidad string    `gorm:"column:tipo_documento_identidad"`
	TipoDocumentoSector    string    `gorm:"column:tipo_documento_sector"`
	Codigoproductosfe      string    `gorm:"column:codigo_producto_sfe"`
	Descripcion            string    `gorm:"column:descripcion" json:"descripcion"`
	SubTotal               float64   `gorm:"column:sub_total"`
	Cantidad               float64   `gorm:"column:cantidad"` // Cambiado a float64
}
type FacturacionDetalles struct {
	ID                    uint      `gorm:"primaryKey;column:id"`
	CreatedBy             string    `gorm:"column:created_by"`
	CreatedDate           time.Time `gorm:"column:created_date"`
	Estado                string    `gorm:"column:estado"`
	ActividadEconomica    string    `gorm:"column:actividad_economica"`
	Cantidad              float64   `gorm:"column:cantidad"` // Cambiado a float64
	CodigoProductoSFE     string    `gorm:"column:codigo_producto_sfe"`
	CodigoProductoSIN     string    `gorm:"column:codigo_producto_sin"`
	Descripcion           string    `gorm:"column:descripcion"`
	PrecioUnitario        float64   `gorm:"column:precio_unitario"`
	SubTotal              float64   `gorm:"column:sub_total"`
	CodigoUnidadMedida    string    `gorm:"column:codigo_unidad_medida"`
	ID_SFEDocumentoFiscal int       `gorm:"column:id_sfe_documento_fiscal"`
	NroItem               int       `gorm:"column:nro_item"`
}

type SFE_factura struct {
	Numero_factura      string `gorm:"column:numero_factura"`
	Cuf                 string `gorm:"column:cuf"`
	Fecha_emision       string `gorm:"column:fecha_emision"`
	Fecha_envio         string `gorm:"column:fecha_envio"`
	Nombre_razon_social string `gorm:"column:nombre_razon_social"`
	Codigo_producto_sfe string `gorm:"column:codigo_producto_sfe"`
	Sub_total           string `gorm:"column:"`
	Precio_unitario     string `gorm:"column:precio_unitario"`
}

// type FacturacionNaabol struct {
// 	ID                          uint                  `gorm:"primaryKey;column:id"`
// 	CreatedBy                   string                `gorm:"column:created_by"`
// 	CreatedDate                 time.Time             `gorm:"column:created_date"`
// 	Estado                      string                `gorm:"column:estado"`
// 	LastModifiedBy              string                `gorm:"column:last_modified_by"`
// 	LastModifiedDate            time.Time             `gorm:"column:last_modified_date"`
// 	Version                     int                   `gorm:"column:version"`
// 	ActividadEconomica          string                `gorm:"column:actividad_economica"`
// 	AjusteAfectacionIVA         float64               `gorm:"column:ajuste_afectacion_iva"`
// 	AjusteNoSujetoIVA           float64               `gorm:"column:ajuste_no_sujeto_iva"`
// 	AjusteSujetoIVA             float64               `gorm:"column:ajuste_sujeto_iva"`
// 	AtributosAdicionales        string                `gorm:"column:atributos_adicionales"`
// 	BeneficiarioLey1886         string                `gorm:"column:beneficiario_ley_1886"`
// 	CAFC                        string                `gorm:"column:cafc"`
// 	CanalEmision                string                `gorm:"column:canal_emision"`
// 	CantidadHabitaciones        int                   `gorm:"column:cantidad_habitaciones"`
// 	CantidadHuespedes           int                   `gorm:"column:cantidad_huespedes"`
// 	CantidadMayores             int                   `gorm:"column:cantidad_mayores"`
// 	CantidadMenores             int                   `gorm:"column:cantidad_menores"`
// 	Ciudad                      string                `gorm:"column:ciudad"`
// 	CodigoActividadEconomica    string                `gorm:"column:codigo_actividad_economica"`
// 	CodigoAutorizacionSC        string                `gorm:"column:codigo_autorizacion_sc"`
// 	CodigoCliente               string                `gorm:"column:codigo_cliente"`
// 	CodigoExcepcion             string                `gorm:"column:codigo_excepcion"`
// 	CodigoIntegracion           string                `gorm:"column:codigo_integracion"`
// 	CodigoRecepcionConjunto     string                `gorm:"column:codigo_recepcion_conjunto"`
// 	CodigoRecepcionSin          string                `gorm:"column:codigo_recepcion_sin"`
// 	CodigoRespuestaSin          string                `gorm:"column:codigo_respuesta_sin"`
// 	CodigoTipoOperacion         string                `gorm:"column:codigo_tipo_operacion"`
// 	CodigosErroresSin           string                `gorm:"column:codigos_errores_sin"`
// 	Complemento                 string                `gorm:"column:complemento"`
// 	CondicionesPago             string                `gorm:"column:condiciones_pago"`
// 	ConsumoPeriodo              float64               `gorm:"column:consumo_periodo"`
// 	CorreoElectronicoCliente    string                `gorm:"column:correo_electronico_cliente"`
// 	CostosGastosInternacionales float64               `gorm:"column:costos_gastos_internacionales"`
// 	CostosGastosNacionales      float64               `gorm:"column:costos_gastos_nacionales"`
// 	CreditoFiscalIVA            float64               `gorm:"column:credito_fiscal_iva"`
// 	CUF                         string                `gorm:"column:cuf"`
// 	CUFD                        string                `gorm:"column:cufd"`
// 	CUIS                        string                `gorm:"column:cuis"`
// 	DebitoFiscalIVA             float64               `gorm:"column:debito_fiscal_iva"`
// 	DescripcionPaquetes         string                `gorm:"column:descripcion_paquetes"`
// 	DescuentoAdicional          float64               `gorm:"column:descuento_adicional"`
// 	DetalleAjusteNoSujetoIVA    string                `gorm:"column:detalle_ajuste_no_sujeto_iva"`
// 	DetalleAjusteSujetoIVA      string                `gorm:"column:detalle_ajuste_sujeto_iva"`
// 	DetaOtrosPagosNoSujetoIVA   string                `gorm:"column:deta_otros_pagos_no_sujeto_iva"`
// 	DireccionSucursal           string                `gorm:"column:direccion_sucursal"`
// 	DomicilioComprador          string                `gorm:"column:domicilio_comprador"`
// 	EstadoDocumentoFiscal       string                `gorm:"column:estado_documento_fiscal"`
// 	Exportacion                 string                `gorm:"column:exportacion"`
// 	FechaEmision                time.Time             `gorm:"column:fecha_emision"`
// 	FechaEmisionFactura         *time.Time            `gorm:"column:fecha_emision_factura"`
// 	FechaEnvio                  *time.Time            `gorm:"column:fecha_envio" json:"fecha_envio"`
// 	FechaIngresoHospedaje       *time.Time            `gorm:"column:fecha_ingreso_hospedaje"`
// 	FechaRespuesta              *time.Time            `gorm:"column:fecha_respuesta"`
// 	Gestion                     int                   `gorm:"column:gestion"`
// 	IDPaquete                   string                `gorm:"column:id_paquete"`
// 	IDTransaccion               string                `gorm:"column:id_transaccion"`
// 	IdentificadorFormato        string                `gorm:"column:identificador_formato"`
// 	IdentificadorURL            string                `gorm:"column:identificador_url"`
// 	Incoterm                    string                `gorm:"column:incoterm"`
// 	IncotermDetalle             string                `gorm:"column:incoterm_detalle"`
// 	InformacionAdicional        string                `gorm:"column:informacion_adicional"`
// 	IngresoDiferenciaCambio     float64               `gorm:"column:ingreso_diferencia_cambio"`
// 	Leyenda                     string                `gorm:"column:leyenda"`
// 	LugarDestino                string                `gorm:"column:lugar_destino"`
// 	Mes                         int                   `gorm:"column:mes"`
// 	Modalidad                   string                `gorm:"column:modalidad"`
// 	ModalidadServicio           string                `gorm:"column:modalidad_servicio"`
// 	MontoDescuento              float64               `gorm:"column:monto_descuento"`
// 	MontoDescuentoLey1886       float64               `gorm:"column:monto_descuento_ley1886"`
// 	MontoDsctoTarifaDignidad    float64               `gorm:"column:monto_dscto_tarifa_dignidad"`
// 	MontoDetalle                float64               `gorm:"column:monto_detalle"`
// 	MontoEfectivoCreditoDebito  float64               `gorm:"column:monto_efectivo_credito_debito"`
// 	MontoGiffCard               float64               `gorm:"column:monto_giff_card"`
// 	MontoICEEspecifico          float64               `gorm:"column:monto_ice_especifico"`
// 	MontoICEPorcentual          float64               `gorm:"column:monto_ice_porcentual"`
// 	MontoIEDH                   float64               `gorm:"column:monto_iedh"`
// 	MontoSeguroInternacional    float64               `gorm:"column:monto_seguro_internacional"`
// 	MontoTotal                  float64               `gorm:"column:monto_total" json:"monto_total"`
// 	MontoTotalArrendFinanciero  float64               `gorm:"column:monto_total_arrend_financiero"`
// 	MontoTotalConciliado        float64               `gorm:"column:monto_total_conciliado"`
// 	MontoTotalDevuelto          float64               `gorm:"column:monto_total_devuelto"`
// 	MontoTotalMoneda            float64               `gorm:"column:monto_total_moneda"`
// 	MontoTotalOriginal          float64               `gorm:"column:monto_total_original"`
// 	MontoTotalSujetoIVA         float64               `gorm:"column:monto_total_sujeto_iva"`
// 	MunicipioDepartamento       string                `gorm:"column:municipio_departamento"`
// 	NITConsolidado              string                `gorm:"column:nit_consolidado"`
// 	NITEmisor                   string                `gorm:"column:nit_emisor"`
// 	NombreEstudiante            string                `gorm:"column:nombre_estudiante"`
// 	NombrePropietario           string                `gorm:"column:nombre_propietario"`
// 	NombreRazonSocial           string                `gorm:"column:nombre_razon_social" json:"nombre_razon_social"`
// 	NombreRepresentanteLegal    string                `gorm:"column:nombre_representante_legal"`
// 	NumeroAutorizacionCUF       string                `gorm:"column:numero_autorizacion_cuf"`
// 	NumeroCelularCliente        string                `gorm:"column:numero_celular_cliente"`
// 	NumeroDocumento             string                `gorm:"column:numero_documento"`
// 	NumeroFactura               string                `gorm:"column:numero_factura"`
// 	NumeroFacturaOriginal       string                `gorm:"column:numero_factura_original"`
// 	NumeroMedidor               string                `gorm:"column:numero_medidor"`
// 	NumeroParteRecepcion        string                `gorm:"column:numero_parte_recepcion"`
// 	NumeroTarjeta               string                `gorm:"column:numero_tarjeta"`
// 	OtrasTasas                  float64               `gorm:"column:otras_tasas"`
// 	OtrosPagosNoSujetoIVA       float64               `gorm:"column:otros_pagos_no_sujeto_iva"`
// 	PeriodoEntrega              string                `gorm:"column:periodo_entrega"`
// 	PeriodoFacturado            string                `gorm:"column:periodo_facturado"`
// 	PlacaVehiculo               string                `gorm:"column:placa_vehiculo"`
// 	PuertoDestino               string                `gorm:"column:puerto_destino"`
// 	RazonSocialEmisor           string                `gorm:"column:razon_social_emisor"`
// 	Regenerate                  string                `gorm:"column:regenerate"`
// 	Sistema                     string                `gorm:"column:sistema"`
// 	TasaAlumbrado               float64               `gorm:"column:tasa_alumbrado"`
// 	TasaAseo                    float64               `gorm:"column:tasa_aseo"`
// 	TelefonoSucursal            string                `gorm:"column:telefono_sucursal"`
// 	TipoCambio                  float64               `gorm:"column:tipo_cambio"`
// 	TipoCambioOficial           float64               `gorm:"column:tipo_cambio_oficial"`
// 	TipoEmision                 string                `gorm:"column:tipo_emision"`
// 	TipoEnvase                  string                `gorm:"column:tipo_envase"`
// 	TipoFactura                 string                `gorm:"column:tipo_factura"`
// 	TotalGastosInternacionales  float64               `gorm:"column:total_gastos_internacionales"`
// 	TotalGastosNacionalesFOB    float64               `gorm:"column:total_gastos_nacionales_fob"`
// 	UsuarioEmision              string                `gorm:"column:usuario_emision"`
// 	Zona                        string                `gorm:"column:zona"`
// 	ZonaSucursal                string                `gorm:"column:zona_sucursal"`
// 	CodigoMoneda                string                `gorm:"column:codigo_moneda"`
// 	CodigoPais                  string                `gorm:"column:codigo_pais"`
// 	IDEvento                    int                   `gorm:"column:id_evento"`
// 	ID_SFECliente               int                   `gorm:"column:id_sfe_cliente"`
// 	ID_SFEPuntoVenta            int                   `gorm:"column:id_sfe_punto_venta"`
// 	ID_SFESucursal              int                   `gorm:"column:id_sfe_sucursal"`
// 	MetodoPago                  string                `gorm:"column:metodo_pago"`
// 	IDReporte                   int                   `gorm:"column:id_reporte"`
// 	TipoDocumentoFiscal         string                `gorm:"column:tipo_documento_fiscal"`
// 	TipoDocumentoIdentidad      string                `gorm:"column:tipo_documento_identidad"`
// 	TipoDocumentoSector         string                `gorm:"column:tipo_documento_sector"`
// 	Codigoproductosfe           string                `gorm:"column:codigo_producto_sfe"`
// 	Descripcion                 string                `gorm:"column:descripcion" json:"descripcion"`
// 	SubTotal                    float64               `gorm:"column:sub_total"`
// 	Cantidad                    float64               `gorm:"column:cantidad"`                  // Cambiado a float64
// 	Detalles                    []FacturacionDetalles `gorm:"foreignKey:ID_SFEDocumentoFiscal"` // Relación uno-a-muchos
// }
// type FacturacionDetalles struct {
// 	ID                          uint      `gorm:"primaryKey;column:id"`
// 	CreatedBy                   string    `gorm:"column:created_by"`
// 	CreatedDate                 time.Time `gorm:"column:created_date"`
// 	Estado                      string    `gorm:"column:estado"`
// 	LastModifiedBy              string    `gorm:"column:last_modified_by"`
// 	LastModifiedDate            time.Time `gorm:"column:last_modified_date"`
// 	Version                     int       `gorm:"column:version"`
// 	ActividadEconomica          string    `gorm:"column:actividad_economica"`
// 	AlicuotaEspecifica          float64   `gorm:"column:alicuota_especifica"`
// 	AlicuotaIVA                 float64   `gorm:"column:alicuota_iva"`
// 	AlicuotaPorcentual          float64   `gorm:"column:alicuota_porcentual"`
// 	AtributosAdicionales        string    `gorm:"column:atributos_adicionales"`
// 	Cantidad                    float64   `gorm:"column:cantidad"` // Cambiado a float64
// 	CantidadICE                 float64   `gorm:"column:cantidad_ice"`
// 	CodigoDetalleTransaccion    string    `gorm:"column:codigo_detalle_transaccion"`
// 	CodigoNandina               string    `gorm:"column:codigo_nandina"`
// 	CodigoProductoSFE           string    `gorm:"column:codigo_producto_sfe"`
// 	CodigoProductoSIN           string    `gorm:"column:codigo_producto_sin"`
// 	CodigoTipoHabitacion        string    `gorm:"column:codigo_tipo_habitacion"`
// 	Descripcion                 string    `gorm:"column:descripcion"`
// 	DetalleHuespedes            string    `gorm:"column:detalle_huespedes"`
// 	Especialidad                string    `gorm:"column:especialidad"`
// 	EspecialidadDetalle         string    `gorm:"column:especialidad_detalle"`
// 	EspecialidadMedico          string    `gorm:"column:especialidad_medico"`
// 	MarcaICE                    string    `gorm:"column:marca_ice"`
// 	MontoConciliado             float64   `gorm:"column:monto_conciliado"`
// 	MontoDescuentoDetalle       float64   `gorm:"column:monto_descuento_detalle"`
// 	MontoFinal                  float64   `gorm:"column:monto_final"`
// 	MontoICEEspecifico          float64   `gorm:"column:monto_ice_especifico"`
// 	MontoICEPorcentual          float64   `gorm:"column:monto_ice_porcentual"`
// 	MontoOriginal               float64   `gorm:"column:monto_original"`
// 	NitDocumentoMedico          string    `gorm:"column:nit_documento_medico"`
// 	NombreApellidoMedico        string    `gorm:"column:nombre_apellido_medico"`
// 	NroFacturaMedico            string    `gorm:"column:nro_factura_medico"`
// 	NroMatriculaMedico          string    `gorm:"column:nro_matricula_medico"`
// 	NroQuirofanoSalaOperaciones string    `gorm:"column:nro_quirofano_sala_operaciones"`
// 	NumeroIMEI                  string    `gorm:"column:numero_imei"`
// 	NumeroSerie                 string    `gorm:"column:numero_serie"`
// 	PorcentajeIEDH              float64   `gorm:"column:porcentaje_iehd"`
// 	PrecioNetoVentaICE          float64   `gorm:"column:precio_neto_venta_ice"`
// 	PrecioUnitario              float64   `gorm:"column:precio_unitario"`
// 	SubTotal                    float64   `gorm:"column:sub_total"`
// 	CodigoUnidadMedida          string    `gorm:"column:codigo_unidad_medida"`
// 	ID_SFEDocumentoFiscal       int       `gorm:"column:id_sfe_documento_fiscal"`
// 	NroItem                     int       `gorm:"column:nro_item"`
// }
