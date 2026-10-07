package services

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"managerfact/internal/domain/models"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type facturadorDatosGenerales struct {
	NitEmisor                   string   `json:"nitEmisor"`
	SucursalEmisor              int      `json:"sucursalEmisor"`
	PuntoVentaEmisor            string   `json:"puntoVentaEmisor"`
	CodigoIntegracion           string   `json:"codigoIntegracion"`
	CodigoCliente               string   `json:"codigoCliente"`
	CelularCliente              *string  `json:"celularCliente"`
	EmailCliente                *string  `json:"emailCliente"`
	AtributosAdicionalesGeneral []string `json:"atributosAdicionalesGeneral"`
}

type facturadorDetalle struct {
	CodigoProducto           string   `json:"codigoProducto"`
	Descripcion              string   `json:"descripcion"`
	Cantidad                 int      `json:"cantidad"`
	PrecioUnitario           float64  `json:"precioUnitario"`
	Subtotal                 float64  `json:"subtotal"`
	MontoDescuentoDetalle    *float64 `json:"montoDescuentoDetalle"`
	CodigoDetalleTransaccion int      `json:"codigoDetalleTransaccion"`
	CodigoUnidadMedida       string   `json:"codigoUnidadMedida"`
}

type facturadorCabecera struct {
	TipoDocumentoFiscal    int     `json:"tipoDocumentoFiscal"`
	TipoDocumentoSector    string  `json:"tipoDocumentoSector"`
	CodigoExcepcion        *string `json:"codigoExcepcion"`
	TipoEmision            int     `json:"tipoEmision"`
	FechaEmision           *string `json:"fechaEmision"`
	NombreRazonSocial      string  `json:"nombreRazonSocial"`
	TipoDocumentoIdentidad string  `json:"tipoDocumentoIdentidad"`
	NumeroDocumento        string  `json:"numeroDocumento"`
	Complemento            *string `json:"complemento"`
	FechaEmisionFactura    *string `json:"fechaEmisionFactura"`
	MetodoPago             string  `json:"metodoPago"`
	CodigoMoneda           string  `json:"codigoMoneda"`
	TipoCambio             float64 `json:"tipoCambio"`
	MontoTotalMoneda       float64 `json:"montoTotalMoneda"`
	MontoTotal             float64 `json:"montoTotal"`
	MontoTotalSujetoIva    float64 `json:"montoTotalSujetoIva"`
	Usuario                string  `json:"usuario"`
}

type facturadorDocumentoFiscal struct {
	Cabecera facturadorCabecera  `json:"cabecera"`
	Detalle  []facturadorDetalle `json:"detalle"`
}

type facturadorRequest struct {
	DatosGenerales  facturadorDatosGenerales  `json:"datosGenerales"`
	DocumentoFiscal facturadorDocumentoFiscal `json:"documentoFiscal"`
}

type FacturadorRespuesta struct {
	Codigo        int    `json:"codigo"`
	Respuesta     string `json:"respuesta"`
	Mensaje       string `json:"mensaje"`
	UrlDocumento  string `json:"urlDocumento"`
	CUF           string `json:"cuf"`
	NumeroFactura int    `json:"numeroFactura"`
}

func redondear2(valor float64) float64 {
	return math.Round(valor*100) / 100
}

// Bolivia usa UTC-4 sin horario de verano.
var zonaLaPaz = time.FixedZone("-04:00", -4*60*60)

// La emisión diferida conserva el día; para hoy resta 10 min porque el facturador rechaza fechas futuras.
func calcularFechaEmision(fechaEmision time.Time) string {
	ahora := time.Now().In(zonaLaPaz)
	var momento time.Time
	if fechaEmision.Year() == ahora.Year() && fechaEmision.Month() == ahora.Month() && fechaEmision.Day() == ahora.Day() {
		momento = ahora.Add(-10 * time.Minute)
	} else {
		momento = time.Date(fechaEmision.Year(), fechaEmision.Month(), fechaEmision.Day(), 10, 0, 0, 0, zonaLaPaz)
	}
	return momento.Format("2006-01-02T15:04:05.000-07:00")
}

func construirPayloadFacturador(factura *models.FacturaPrevalorada, sucursal *models.SucursalFacturador) facturadorRequest {

	montoTotal := factura.TotalBob
	fechaEmision := calcularFechaEmision(factura.FechaEmision)

	return facturadorRequest{
		DatosGenerales: facturadorDatosGenerales{

			NitEmisor:                   "419945029",
			SucursalEmisor:              sucursal.CodigoSucursalSin,
			PuntoVentaEmisor:            sucursal.PuntoVentaEmisor,
			CodigoIntegracion:           factura.CodigoIntegracion,
			CodigoCliente:               "N/A",
			CelularCliente:              nil,
			EmailCliente:                nil,
			AtributosAdicionalesGeneral: []string{},
		},
		DocumentoFiscal: facturadorDocumentoFiscal{
			Cabecera: facturadorCabecera{
				TipoDocumentoFiscal:    1,
				TipoDocumentoSector:    "23",
				CodigoExcepcion:        nil,
				TipoEmision:            3,
				FechaEmision:           &fechaEmision,
				NombreRazonSocial:      "S/N",
				TipoDocumentoIdentidad: "NIT",
				NumeroDocumento:        "0",
				Complemento:            nil,
				FechaEmisionFactura:    nil,
				MetodoPago:             "1",
				CodigoMoneda:           sucursal.CodigoMonedaBob,
				TipoCambio:             1,
				MontoTotalMoneda:       montoTotal,
				MontoTotal:             montoTotal,
				MontoTotalSujetoIva:    montoTotal,
				Usuario:                "ManagerFact",
			},
			Detalle: []facturadorDetalle{
				{
					CodigoProducto:           factura.CodigoProducto,
					Descripcion:              factura.Detalle,
					Cantidad:                 1,
					PrecioUnitario:           montoTotal,
					Subtotal:                 montoTotal,
					MontoDescuentoDetalle:    nil,
					CodigoDetalleTransaccion: 1,
					CodigoUnidadMedida:       "58",
				},
			},
		},
	}
}

var httpClienteFacturador = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func enviarAFacturador(sucursal *models.SucursalFacturador, factura *models.FacturaPrevalorada, tokenAcceso string) (*FacturadorRespuesta, error) {
	payload := construirPayloadFacturador(factura, sucursal)
	url := strings.TrimRight(sucursal.UrlLinkFacturador, "/") + "/clic-core/facturas/recibir-sincrono"
	return postFacturador(url, payload, tokenAcceso, factura.CodigoIntegracion)
}

func postFacturador(url string, payload any, tokenAcceso string, codigoIntegracion string) (*FacturadorRespuesta, error) {
	cuerpo, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error armando el JSON de la request: %w", err)
	}
	log.Printf("[FacturadorClient] POST %s codigo_integracion=%s payload=%s", url, codigoIntegracion, cuerpo)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(cuerpo))
	if err != nil {
		return nil, fmt.Errorf("error creando la request al facturador: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenAcceso)

	resp, err := httpClienteFacturador.Do(req)
	if err != nil {
		log.Printf("[FacturadorClient] error de transporte codigo_integracion=%s: %v", codigoIntegracion, err)
		return nil, fmt.Errorf("error llamando al facturador: %w", err)
	}
	defer resp.Body.Close()

	cuerpoResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo la respuesta del facturador: %w", err)
	}
	log.Printf("[FacturadorClient] respuesta status=%d codigo_integracion=%s body=%s", resp.StatusCode, codigoIntegracion, cuerpoResp)

	var respuesta FacturadorRespuesta
	if err := json.Unmarshal(cuerpoResp, &respuesta); err != nil {
		return nil, fmt.Errorf("respuesta del facturador no es el JSON esperado (status %d, body %q): %w", resp.StatusCode, cuerpoResp, err)
	}

	if respuesta.Mensaje == "" {
		respuesta.Mensaje = fmt.Sprintf("(status %d) %s", resp.StatusCode, string(cuerpoResp))
	}
	return &respuesta, nil
}

type anulacionDatosGenerales struct {
	NitEmisor        json.Number `json:"nitEmisor"`
	SucursalEmisor   string      `json:"sucursalEmisor"`
	PuntoVentaEmisor string      `json:"puntoVentaEmisor"`
	CanalFacturacion string      `json:"canalFacturacion"`
}

type anulacionDocumentoFiscal struct {
	CodigoIntegracion string `json:"codigoIntegracion"`
	Cuf               string `json:"cuf"`
	CodigoMotivo      string `json:"codigoMotivo"`
}

type anulacionRequest struct {
	DatosGenerales  anulacionDatosGenerales  `json:"datosGenerales"`
	DocumentoFiscal anulacionDocumentoFiscal `json:"documentoFiscal"`
}

func construirPayloadAnulacion(factura *models.FacturaAnulacion, sucursal *models.SucursalFacturador) anulacionRequest {
	return anulacionRequest{
		DatosGenerales: anulacionDatosGenerales{
			NitEmisor:        "419945029",
			SucursalEmisor:   strconv.Itoa(sucursal.CodigoSucursalSin),
			PuntoVentaEmisor: sucursal.PuntoVentaEmisor,
			CanalFacturacion: "core",
		},
		DocumentoFiscal: anulacionDocumentoFiscal{
			CodigoIntegracion: factura.CodigoIntegracion,
			Cuf:               factura.Cuf,
			CodigoMotivo:      factura.CodigoMotivo,
		},
	}
}

func enviarAAnular(sucursal *models.SucursalFacturador, factura *models.FacturaAnulacion, tokenAcceso string) (*FacturadorRespuesta, error) {
	payload := construirPayloadAnulacion(factura, sucursal)
	url := strings.TrimRight(sucursal.UrlLinkFacturador, "/") + "/clic-core/facturas/anular"
	return postFacturador(url, payload, tokenAcceso, factura.CodigoIntegracion)
}
