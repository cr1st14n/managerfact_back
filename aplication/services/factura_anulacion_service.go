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

// ErrAnulacionYaAceptada se devuelve al intentar anular un registro que el
// facturador ya aceptó — reenviarlo no tiene efecto y solo generaría ruido.
var ErrAnulacionYaAceptada = errors.New("esta anulación ya fue aceptada por el facturador, no se puede reenviar")

type FacturaAnulacionService struct {
	repo               *repositories.FacturaAnulacionRepository
	sucursalFacturador *repositories.SucursalFacturadorRepository
	logEnvio           *repositories.LogEnvioRepository
	usuarioService     *UsuarioService
}

func NewFacturaAnulacionService(
	r *repositories.FacturaAnulacionRepository,
	sucursalFacturadorRepo *repositories.SucursalFacturadorRepository,
	logEnvioRepo *repositories.LogEnvioRepository,
	usuarioService *UsuarioService,
) *FacturaAnulacionService {
	return &FacturaAnulacionService{repo: r, sucursalFacturador: sucursalFacturadorRepo, logEnvio: logEnvioRepo, usuarioService: usuarioService}
}

// columnasEsperadasAnulacion son los encabezados de columna del Excel de
// anulación (ver doc/EnvioFacturacion.md sección 4). codigo_integracion es
// el de la factura original a anular — a diferencia de la prevalorada, acá
// no se genera: viene del Excel.
var columnasEsperadasAnulacion = []string{"cuf", "codigo_motivo", "codigo_integracion"}

// ImportarExcelAnulacion parsea un archivo .xlsx de anulaciones y guarda las
// filas válidas como facturas_anulacion en estado "pendiente", todas
// fijadas a la sucursalFacturadorID elegida antes de importar. Las filas
// inválidas se reportan pero no abortan el archivo completo.
func (s *FacturaAnulacionService) ImportarExcel(usuarioID uint, archivo io.Reader, sucursalFacturadorID uint, observacion string) (*ImportarExcelResultado, error) {
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
	for _, columna := range columnasEsperadasAnulacion {
		if _, ok := indiceColumna[columna]; !ok {
			return nil, fmt.Errorf("falta la columna requerida %q en el Excel", columna)
		}
	}

	loteID := uuid.NewString()
	validas := []models.FacturaAnulacion{}
	conError := []FilaConError{}

	for i, fila := range filas[1:] {
		numeroFila := i + 2 // +1 por índice base 0, +1 por la fila de encabezado
		factura, err := parsearFilaAnulacion(fila, indiceColumna, sucursalFacturadorID, loteID, observacion)
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

func parsearFilaAnulacion(fila []string, indiceColumna map[string]int, sucursalFacturadorID uint, loteID string, observacion string) (*models.FacturaAnulacion, error) {
	cuf := valorColumna(fila, indiceColumna, "cuf")
	codigoMotivo := valorColumna(fila, indiceColumna, "codigo_motivo")
	codigoIntegracion := valorColumna(fila, indiceColumna, "codigo_integracion")

	if cuf == "" {
		return nil, fmt.Errorf("cuf es requerido")
	}
	if codigoMotivo == "" {
		return nil, fmt.Errorf("codigo_motivo es requerido")
	}
	if codigoIntegracion == "" {
		return nil, fmt.Errorf("codigo_integracion es requerido")
	}

	return &models.FacturaAnulacion{
		SucursalFacturadorID: sucursalFacturadorID,
		LoteID:               loteID,
		CodigoIntegracion:    codigoIntegracion,
		Observacion:          observacion,
		Cuf:                  cuf,
		CodigoMotivo:         codigoMotivo,
		Estado:               "pendiente",
	}, nil
}

func (s *FacturaAnulacionService) ObtenerPorID(usuarioID, id uint) (*models.FacturaAnulacion, error) {
	factura, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if factura.SucursalFacturador == nil {
		return nil, fmt.Errorf("la sucursal facturador de esta anulación no existe o fue eliminada")
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

// Anular arma el JSON de la solicitud de anulación (CUF + motivo + sucursal
// facturador) y lo envía a clic-core/facturas/anular (ver
// doc/EnvioFacturacion.md secciones 4 y 5). Guarda el resultado del intento
// (aceptado/rechazado/error) incluso si la llamada falla, para no perder el
// rastro del envío. origen es "manual" (botón del front) o "automatico"
// (EnvioWorker) — solo se usa para el registro en logs_envio.
func (s *FacturaAnulacionService) Anular(id uint, origen string) (*models.FacturaAnulacion, error) {
	factura, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if factura.Estado == "aceptado" {
		return nil, ErrAnulacionYaAceptada
	}
	if factura.SucursalFacturador == nil {
		return nil, fmt.Errorf("la sucursal facturador de esta anulación no existe o fue eliminada")
	}

	tokenAcceso, err := utils.Decrypt(factura.SucursalFacturador.TokenAcceso)
	if err != nil {
		return nil, fmt.Errorf("error descifrando el token de la sucursal facturador: %w", err)
	}

	ahora := time.Now()
	factura.FechaEnvio = &ahora
	factura.Estado = "enviado"

	respuesta, err := enviarAAnular(factura.SucursalFacturador, factura, tokenAcceso)
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
			log.Printf("[FacturaAnulacionService] error marcando sucursal %d en_revision: %v", factura.SucursalFacturadorID, marcarErr)
		}
		s.registrarLog(factura.ID, factura.CodigoIntegracion, factura.SucursalFacturadorID, origen, "error", err.Error())
		return factura, fmt.Errorf("error enviando la anulación al facturador: %w", err)
	}

	// El facturador respondió (aceptado o rechazado): la sucursal está
	// alcanzable, así que si estaba "en_revision" se recupera sola.
	if factura.SucursalFacturador.EstadoConexion == "en_revision" {
		if marcarErr := s.sucursalFacturador.ActualizarEstadoConexion(factura.SucursalFacturadorID, "activo", "", nil); marcarErr != nil {
			log.Printf("[FacturaAnulacionService] error marcando sucursal %d activa: %v", factura.SucursalFacturadorID, marcarErr)
		}
	}

	factura.CodigoRespuesta = strconv.Itoa(respuesta.Codigo)
	factura.MensajeRespuesta = respuesta.Mensaje
	if respuesta.Codigo == 200 && respuesta.Respuesta == "OK" {
		factura.Estado = "aceptado"
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
// el flujo de anulación, solo se loguea a consola.
func (s *FacturaAnulacionService) registrarLog(facturaID uint, codigoIntegracion string, sucursalFacturadorID uint, origen, resultado, mensaje string) {
	entrada := &models.LogEnvio{
		Tipo:                 "anulacion",
		FacturaID:            facturaID,
		CodigoIntegracion:    codigoIntegracion,
		SucursalFacturadorID: sucursalFacturadorID,
		Origen:               origen,
		Resultado:            resultado,
		Mensaje:              mensaje,
	}
	if err := s.logEnvio.Create(entrada); err != nil {
		log.Printf("[FacturaAnulacionService] error guardando log de envío: %v", err)
	}
}

// ListarTodos devuelve solo las facturas de anulación de sucursales que el
// usuario tiene permitidas (ver codigosSucursalPermitidos).
func (s *FacturaAnulacionService) ListarTodos(usuarioID uint, estado, loteID string) ([]models.FacturaAnulacion, error) {
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
	visibles := make([]models.FacturaAnulacion, 0, len(facturas))
	for _, f := range facturas {
		if f.SucursalFacturador != nil && permitidos[f.SucursalFacturador.CodigoSucursalSin] {
			visibles = append(visibles, f)
		}
	}
	return visibles, nil
}

// ListarPendientesParaEnvio expone las facturas de anulación pendientes
// para el EnvioWorker, en el orden en que deben procesarse.
func (s *FacturaAnulacionService) ListarPendientesParaEnvio() ([]models.FacturaAnulacion, error) {
	return s.repo.GetPendientesParaEnvio()
}

// ListarLotes agrega las facturas de anulación por lote de importación,
// filtrando a las sucursales permitidas del usuario; el detalle de cada
// lote se obtiene después con ListarTodos(usuarioID, "", loteID).
func (s *FacturaAnulacionService) ListarLotes(usuarioID uint) ([]repositories.LoteResumenAnulacion, error) {
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
	visibles := make([]repositories.LoteResumenAnulacion, 0, len(lotes))
	for _, l := range lotes {
		if permitidos[l.CodigoSucursalSin] {
			visibles = append(visibles, l)
		}
	}
	return visibles, nil
}

// GenerarPlantilla arma el .xlsx de ejemplo con las columnas que espera
// ImportarExcel, para que el usuario sepa en qué formato cargar el archivo.
func (s *FacturaAnulacionService) GenerarPlantilla() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	hoja := f.GetSheetName(0)

	for i, columna := range columnasEsperadasAnulacion {
		celda, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(hoja, celda, columna)
	}

	ejemplo := []any{"E229797000005010004070100000000120250716212B4198F4", "1", "1025000001832577"}
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
