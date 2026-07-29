package services

import (
	"fmt"
	"io"
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type FacturaAnulacionService struct {
	repo               *repositories.FacturaAnulacionRepository
	sucursalFacturador *repositories.SucursalFacturadorRepository
}

func NewFacturaAnulacionService(
	r *repositories.FacturaAnulacionRepository,
	sucursalFacturadorRepo *repositories.SucursalFacturadorRepository,
) *FacturaAnulacionService {
	return &FacturaAnulacionService{repo: r, sucursalFacturador: sucursalFacturadorRepo}
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
func (s *FacturaAnulacionService) ImportarExcel(archivo io.Reader, sucursalFacturadorID uint, observacion string) (*ImportarExcelResultado, error) {
	if _, err := s.sucursalFacturador.GetByID(sucursalFacturadorID); err != nil {
		return nil, fmt.Errorf("sucursal facturador inválida: %w", err)
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

func (s *FacturaAnulacionService) ObtenerPorID(id uint) (*models.FacturaAnulacion, error) {
	return s.repo.GetByID(id)
}

func (s *FacturaAnulacionService) ListarTodos(estado, loteID string) ([]models.FacturaAnulacion, error) {
	return s.repo.GetAll(estado, loteID)
}

// ListarLotes agrega las facturas de anulación por lote de importación; el
// detalle de cada lote se obtiene después con ListarTodos("", loteID).
func (s *FacturaAnulacionService) ListarLotes() ([]repositories.LoteResumenAnulacion, error) {
	return s.repo.GetLotes()
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
