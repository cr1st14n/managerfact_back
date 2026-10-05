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

var columnasEsperadasAnulacion = []string{"cuf", "codigo_motivo", "codigo_integracion"}

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
		numeroFila := i + 2 
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

		if marcarErr := s.sucursalFacturador.ActualizarEstadoConexion(factura.SucursalFacturadorID, "en_revision", err.Error(), &fechaRespuesta); marcarErr != nil {
			log.Printf("[FacturaAnulacionService] error marcando sucursal %d en_revision: %v", factura.SucursalFacturadorID, marcarErr)
		}
		s.registrarLog(factura.ID, factura.CodigoIntegracion, factura.SucursalFacturadorID, origen, "error", err.Error())
		return factura, fmt.Errorf("error enviando la anulación al facturador: %w", err)
	}

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

func (s *FacturaAnulacionService) ListarPendientesParaEnvio() ([]models.FacturaAnulacion, error) {
	return s.repo.GetPendientesParaEnvio()
}

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
