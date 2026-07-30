package services

import (
	"errors"
	"fmt"
	"io"
	"log"
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"
	"managerfact/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// ErrFacturaYaAceptada se devuelve al intentar facturar un registro que el
// facturador ya aceptó — reenviarlo generaría un documento fiscal duplicado.
var ErrFacturaYaAceptada = errors.New("esta factura ya fue aceptada por el facturador, no se puede reenviar")

// ErrSinPermisoSucursal se devuelve cuando el usuario autenticado intenta
// cargar un Excel o consultar facturas de una sucursal que no tiene entre
// sus sucursales permitidas (usuarios.sucursales_permitidas_codigos).
var ErrSinPermisoSucursal = errors.New("no tienes permiso para esta sucursal")

type FacturaPrevaloradaService struct {
	repo               *repositories.FacturaPrevaloradaRepository
	sucursalFacturador *repositories.SucursalFacturadorRepository
	logEnvio           *repositories.LogEnvioRepository
	usuarioService     *UsuarioService
}

func NewFacturaPrevaloradaService(
	r *repositories.FacturaPrevaloradaRepository,
	sucursalFacturadorRepo *repositories.SucursalFacturadorRepository,
	logEnvioRepo *repositories.LogEnvioRepository,
	usuarioService *UsuarioService,
) *FacturaPrevaloradaService {
	return &FacturaPrevaloradaService{repo: r, sucursalFacturador: sucursalFacturadorRepo, logEnvio: logEnvioRepo, usuarioService: usuarioService}
}

// codigosSucursalPermitidos resuelve, para el conjunto de codigo_sucursal_sin
// dado, cuáles puede ver el usuario — usa el mismo chequeo directo que
// TieneAccesoSucursal (acceso_total o coincidencia en
// sucursales_permitidas_codigos) para cada código distinto, SIN pasar por el
// catálogo sucursales_catalogo: SucursalFacturador es independiente de ese
// catálogo (ver doc/EnvioFacturacion.md sección 1), así que un código de
// sucursal facturador que no esté en el catálogo igual debe ser visible si
// el usuario tiene permiso sobre ese código.
func codigosSucursalPermitidos(usuarioService *UsuarioService, usuarioID uint, codigos []int) (map[int]bool, error) {
	permitidos := make(map[int]bool, len(codigos))
	vistos := make(map[int]bool, len(codigos))
	for _, codigo := range codigos {
		if vistos[codigo] {
			continue
		}
		vistos[codigo] = true
		ok, err := usuarioService.TieneAccesoSucursal(usuarioID, codigo)
		if err != nil {
			return nil, fmt.Errorf("error verificando accesos: %w", err)
		}
		if ok {
			permitidos[codigo] = true
		}
	}
	return permitidos, nil
}

// columnasEsperadas son los encabezados de columna del Excel de boletos
// (ver doc/EnvioFacturacion.md sección 3).
var columnasEsperadas = []string{
	"detalle", "costo_dua_dolares", "fecha_emision",
	"fecha_compra_boleto", "tipo_cambio", "codigo_producto",
}

// FilaConError describe una fila del Excel que no se pudo importar.
type FilaConError struct {
	Fila   int    `json:"fila"`
	Motivo string `json:"motivo"`
}

// ImportarExcelResultado es la respuesta del endpoint de importación.
type ImportarExcelResultado struct {
	LoteID   string         `json:"lote_id"`
	Total    int            `json:"total"`
	Validas  int            `json:"validas"`
	ConError []FilaConError `json:"con_error"`
}

// ImportarExcel parsea un archivo .xlsx de boletos y guarda las filas
// válidas como facturas_prevaloradas en estado "pendiente", todas fijadas a
// la sucursalFacturadorID elegida antes de importar (etapa 1 del flujo).
// Las filas inválidas se reportan pero no abortan el archivo completo.
func (s *FacturaPrevaloradaService) ImportarExcel(usuarioID uint, archivo io.Reader, sucursalFacturadorID uint, observacion string) (*ImportarExcelResultado, error) {
	sucursal, err := s.sucursalFacturador.GetByID(sucursalFacturadorID)
	if err != nil {
		return nil, fmt.Errorf("sucursal facturador inválida: %w", err)
	}
	permitido, err := s.usuarioService.TieneAccesoSucursal(usuarioID, sucursal.CodigoSucursalSin)
	if err != nil {
		return nil, fmt.Errorf("error verificando accesos: %w", err)
	}
	if !permitido {
		return nil, ErrSinPermisoSucursal
	}
	observacion = strings.TrimSpace(observacion)
	if observacion == "" {
		return nil, fmt.Errorf("observacion es requerida: indica el motivo de carga del lote")
	}

	f, err := excelize.OpenReader(archivo)
	if err != nil {
		return nil, fmt.Errorf("archivo Excel inválido: %w", err)
	}
	defer f.Close()

	hojas := f.GetSheetList()
	if len(hojas) == 0 {
		return nil, fmt.Errorf("el archivo Excel no tiene hojas")
	}

	filas, err := f.GetRows(hojas[0])
	if err != nil {
		return nil, fmt.Errorf("error leyendo la hoja del Excel: %w", err)
	}
	if len(filas) < 2 {
		return nil, fmt.Errorf("el archivo Excel no tiene filas de datos")
	}

	indiceColumna := mapearColumnas(filas[0])
	for _, columna := range columnasEsperadas {
		if _, ok := indiceColumna[columna]; !ok {
			return nil, fmt.Errorf("falta la columna requerida %q en el Excel", columna)
		}
	}

	loteID := uuid.NewString()
	validas := []models.FacturaPrevalorada{}
	conError := []FilaConError{}

	for i, fila := range filas[1:] {
		numeroFila := i + 2 // +1 por índice base 0, +1 por la fila de encabezado
		factura, err := parsearFilaBoleto(fila, indiceColumna, sucursalFacturadorID, loteID, observacion)
		if err != nil {
			conError = append(conError, FilaConError{Fila: numeroFila, Motivo: err.Error()})
			continue
		}
		validas = append(validas, *factura)
	}

	if err := s.repo.CreateBatch(validas); err != nil {
		return nil, err
	}

	return &ImportarExcelResultado{
		LoteID:   loteID,
		Total:    len(filas) - 1,
		Validas:  len(validas),
		ConError: conError,
	}, nil
}

func mapearColumnas(encabezados []string) map[string]int {
	indice := make(map[string]int, len(encabezados))
	for i, encabezado := range encabezados {
		clave := strings.ToLower(strings.TrimSpace(encabezado))
		indice[clave] = i
	}
	return indice
}

func valorColumna(fila []string, indiceColumna map[string]int, columna string) string {
	idx, ok := indiceColumna[columna]
	if !ok || idx >= len(fila) {
		return ""
	}
	return strings.TrimSpace(fila[idx])
}

func parsearFilaBoleto(fila []string, indiceColumna map[string]int, sucursalFacturadorID uint, loteID string, observacion string) (*models.FacturaPrevalorada, error) {
	detalle := valorColumna(fila, indiceColumna, "detalle")
	codigoProducto := valorColumna(fila, indiceColumna, "codigo_producto")
	costoDuaStr := valorColumna(fila, indiceColumna, "costo_dua_dolares")
	fechaEmisionStr := valorColumna(fila, indiceColumna, "fecha_emision")
	fechaCompraBoletoStr := valorColumna(fila, indiceColumna, "fecha_compra_boleto")
	tipoCambioStr := valorColumna(fila, indiceColumna, "tipo_cambio")

	if detalle == "" {
		return nil, fmt.Errorf("detalle es requerido")
	}
	if codigoProducto == "" {
		return nil, fmt.Errorf("codigo_producto es requerido")
	}

	costoDua, err := strconv.ParseFloat(costoDuaStr, 64)
	if err != nil {
		return nil, fmt.Errorf("costo_dua_dolares inválido: %q", costoDuaStr)
	}

	tipoCambio, err := strconv.ParseFloat(tipoCambioStr, 64)
	if err != nil {
		return nil, fmt.Errorf("tipo_cambio inválido: %q", tipoCambioStr)
	}

	fechaEmision, err := parsearFecha(fechaEmisionStr)
	if err != nil {
		return nil, fmt.Errorf("fecha_emision inválida: %q", fechaEmisionStr)
	}

	fechaCompraBoleto, err := parsearFecha(fechaCompraBoletoStr)
	if err != nil {
		return nil, fmt.Errorf("fecha_compra_boleto inválida: %q", fechaCompraBoletoStr)
	}

	return &models.FacturaPrevalorada{
		SucursalFacturadorID: sucursalFacturadorID,
		LoteID:               loteID,
		CodigoIntegracion:    uuid.NewString(),
		Tipo:                 "FACTURA_PREVALORADA",
		Observacion:          observacion,
		Detalle:              detalle,
		CodigoProducto:       codigoProducto,
		CostoDuaDolares:      costoDua,
		FechaCompraBoleto:    fechaCompraBoleto,
		TipoCambio:           tipoCambio,
		TotalBob:             redondear2(costoDua * tipoCambio),
		FechaEmision:         fechaEmision,
		Estado:               "pendiente",
	}, nil
}

var formatosFecha = []string{"2006-01-02", "02/01/2006", "2/1/2006"}

func parsearFecha(valor string) (time.Time, error) {
	if valor == "" {
		return time.Time{}, fmt.Errorf("valor vacío")
	}
	for _, formato := range formatosFecha {
		if fecha, err := time.Parse(formato, valor); err == nil {
			return fecha, nil
		}
	}
	// Excel puede entregar la fecha como número de serie (celda con
	// formato de fecha real en vez de texto).
	if serie, err := strconv.ParseFloat(valor, 64); err == nil {
		if fecha, err := excelize.ExcelDateToTime(serie, false); err == nil {
			return fecha, nil
		}
	}
	return time.Time{}, fmt.Errorf("formato de fecha no reconocido")
}

func (s *FacturaPrevaloradaService) ObtenerPorID(usuarioID, id uint) (*models.FacturaPrevalorada, error) {
	factura, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if factura.SucursalFacturador == nil {
		return nil, fmt.Errorf("la sucursal facturador de esta factura no existe o fue eliminada")
	}
	permitido, err := s.usuarioService.TieneAccesoSucursal(usuarioID, factura.SucursalFacturador.CodigoSucursalSin)
	if err != nil {
		return nil, fmt.Errorf("error verificando accesos: %w", err)
	}
	if !permitido {
		return nil, ErrSinPermisoSucursal
	}
	return factura, nil
}

// Facturar arma el JSON de la factura prevalorada (boleto + sucursal
// facturador) y lo envía a clic-core/facturas/recibir-sincrono (etapa 2 del
// flujo, ver doc/EnvioFacturacion.md sección 2 y 5). Guarda el resultado del
// intento (aceptado/rechazado/error) incluso si la llamada falla, para no
// perder el rastro del envío. origen es "manual" (botón del front) o
// "automatico" (EnvioWorker) — solo se usa para el registro en logs_envio.
func (s *FacturaPrevaloradaService) Facturar(id uint, origen string) (*models.FacturaPrevalorada, error) {
	factura, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if factura.Estado == "aceptado" {
		return nil, ErrFacturaYaAceptada
	}
	if factura.SucursalFacturador == nil {
		return nil, fmt.Errorf("la sucursal facturador de esta factura no existe o fue eliminada")
	}

	tokenAcceso, err := utils.Decrypt(factura.SucursalFacturador.TokenAcceso)
	if err != nil {
		return nil, fmt.Errorf("error descifrando el token de la sucursal facturador: %w", err)
	}

	ahora := time.Now()
	factura.FechaEnvio = &ahora
	factura.Estado = "enviado"

	respuesta, err := enviarAFacturador(factura.SucursalFacturador, factura, tokenAcceso)
	fechaRespuesta := time.Now()
	factura.FechaRespuesta = &fechaRespuesta

	if err != nil {
		factura.Estado = "error"
		factura.MensajeRespuesta = err.Error()
		if guardarErr := s.repo.Update(factura); guardarErr != nil {
			return nil, guardarErr
		}
		// Error de transporte (no de negocio): la sucursal facturador queda
		// "en_revision" para que el EnvioWorker deje de insistir con ella
		// hasta que vuelva a responder — ver doc/EnvioFacturacion.md sección 5.
		if marcarErr := s.sucursalFacturador.ActualizarEstadoConexion(factura.SucursalFacturadorID, "en_revision", err.Error(), &fechaRespuesta); marcarErr != nil {
			log.Printf("[FacturaPrevaloradaService] error marcando sucursal %d en_revision: %v", factura.SucursalFacturadorID, marcarErr)
		}
		s.registrarLog(factura.ID, factura.CodigoIntegracion, factura.SucursalFacturadorID, origen, "error", err.Error())
		return factura, fmt.Errorf("error enviando al facturador: %w", err)
	}

	// El facturador respondió (aceptado o rechazado): la sucursal está
	// alcanzable, así que si estaba "en_revision" se recupera sola.
	if factura.SucursalFacturador.EstadoConexion == "en_revision" {
		if marcarErr := s.sucursalFacturador.ActualizarEstadoConexion(factura.SucursalFacturadorID, "activo", "", nil); marcarErr != nil {
			log.Printf("[FacturaPrevaloradaService] error marcando sucursal %d activa: %v", factura.SucursalFacturadorID, marcarErr)
		}
	}

	factura.CodigoRespuesta = strconv.Itoa(respuesta.Codigo)
	factura.MensajeRespuesta = respuesta.Mensaje
	if respuesta.Codigo == 200 && respuesta.Respuesta == "OK" {
		factura.Estado = "aceptado"
		factura.CUF = respuesta.CUF
		factura.UrlDocumento = respuesta.UrlDocumento
		if respuesta.NumeroFactura != 0 {
			factura.NumeroFactura = strconv.Itoa(respuesta.NumeroFactura)
		}
	} else {
		factura.Estado = "rechazado"
		factura.MensajeRespuesta = fmt.Sprintf("rechazado: %s", respuesta.Mensaje)
	}

	if err := s.repo.Update(factura); err != nil {
		return nil, err
	}
	s.registrarLog(factura.ID, factura.CodigoIntegracion, factura.SucursalFacturadorID, origen, factura.Estado, factura.MensajeRespuesta)
	return factura, nil
}

// registrarLog guarda el intento en logs_envio; un fallo acá no debe abortar
// el flujo de facturación, solo se loguea a consola.
func (s *FacturaPrevaloradaService) registrarLog(facturaID uint, codigoIntegracion string, sucursalFacturadorID uint, origen, resultado, mensaje string) {
	entrada := &models.LogEnvio{
		Tipo:                 "prevalorada",
		FacturaID:            facturaID,
		CodigoIntegracion:    codigoIntegracion,
		SucursalFacturadorID: sucursalFacturadorID,
		Origen:               origen,
		Resultado:            resultado,
		Mensaje:              mensaje,
	}
	if err := s.logEnvio.Create(entrada); err != nil {
		log.Printf("[FacturaPrevaloradaService] error guardando log de envío: %v", err)
	}
}

// ListarTodos devuelve solo las facturas de sucursales que el usuario tiene
// permitidas (ver codigosSucursalPermitidos).
func (s *FacturaPrevaloradaService) ListarTodos(usuarioID uint, estado, loteID string) ([]models.FacturaPrevalorada, error) {
	facturas, err := s.repo.GetAll(estado, loteID)
	if err != nil {
		return nil, err
	}
	codigos := make([]int, 0, len(facturas))
	for _, f := range facturas {
		if f.SucursalFacturador != nil {
			codigos = append(codigos, f.SucursalFacturador.CodigoSucursalSin)
		}
	}
	permitidos, err := codigosSucursalPermitidos(s.usuarioService, usuarioID, codigos)
	if err != nil {
		return nil, err
	}
	visibles := make([]models.FacturaPrevalorada, 0, len(facturas))
	for _, f := range facturas {
		if f.SucursalFacturador != nil && permitidos[f.SucursalFacturador.CodigoSucursalSin] {
			visibles = append(visibles, f)
		}
	}
	return visibles, nil
}

// ListarPendientesParaEnvio expone las facturas pendientes para el
// EnvioWorker, en el orden en que deben procesarse.
func (s *FacturaPrevaloradaService) ListarPendientesParaEnvio() ([]models.FacturaPrevalorada, error) {
	return s.repo.GetPendientesParaEnvio()
}

// ListarLotes agrega las facturas por lote de importación (registro de
// lotes), filtrando a las sucursales permitidas del usuario; el detalle de
// cada lote se obtiene después con ListarTodos(usuarioID, "", loteID).
func (s *FacturaPrevaloradaService) ListarLotes(usuarioID uint) ([]repositories.LoteResumen, error) {
	lotes, err := s.repo.GetLotes()
	if err != nil {
		return nil, err
	}
	codigos := make([]int, 0, len(lotes))
	for _, l := range lotes {
		codigos = append(codigos, l.CodigoSucursalSin)
	}
	permitidos, err := codigosSucursalPermitidos(s.usuarioService, usuarioID, codigos)
	if err != nil {
		return nil, err
	}
	visibles := make([]repositories.LoteResumen, 0, len(lotes))
	for _, l := range lotes {
		if permitidos[l.CodigoSucursalSin] {
			visibles = append(visibles, l)
		}
	}
	return visibles, nil
}

// GenerarPlantilla arma el .xlsx de ejemplo con las columnas que espera
// ImportarExcel, para que el usuario sepa en qué formato cargar el archivo.
func (s *FacturaPrevaloradaService) GenerarPlantilla() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	hoja := f.GetSheetName(0)

	for i, columna := range columnasEsperadas {
		celda, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(hoja, celda, columna)
	}

	ejemplo := []any{
		"Uso de sala aeropuerto",
		13.90,
		"2026-01-15",
		"2026-01-10",
		6.96,
		"99101",
	}
	for i, valor := range ejemplo {
		celda, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(hoja, celda, valor)
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("error generando plantilla: %w", err)
	}
	return buffer.Bytes(), nil
}
