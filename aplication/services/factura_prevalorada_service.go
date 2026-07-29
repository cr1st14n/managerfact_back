package services

import (
	"fmt"
	"io"
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type FacturaPrevaloradaService struct {
	repo               *repositories.FacturaPrevaloradaRepository
	sucursalFacturador *repositories.SucursalFacturadorRepository
}

func NewFacturaPrevaloradaService(
	r *repositories.FacturaPrevaloradaRepository,
	sucursalFacturadorRepo *repositories.SucursalFacturadorRepository,
) *FacturaPrevaloradaService {
	return &FacturaPrevaloradaService{repo: r, sucursalFacturador: sucursalFacturadorRepo}
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
func (s *FacturaPrevaloradaService) ImportarExcel(archivo io.Reader, sucursalFacturadorID uint, observacion string) (*ImportarExcelResultado, error) {
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

func (s *FacturaPrevaloradaService) ObtenerPorID(id uint) (*models.FacturaPrevalorada, error) {
	return s.repo.GetByID(id)
}

func (s *FacturaPrevaloradaService) ListarTodos(estado, loteID string) ([]models.FacturaPrevalorada, error) {
	return s.repo.GetAll(estado, loteID)
}

// ListarLotes agrega las facturas por lote de importación (registro de
// lotes); el detalle de cada lote se obtiene después con
// ListarTodos("", loteID).
func (s *FacturaPrevaloradaService) ListarLotes() ([]repositories.LoteResumen, error) {
	return s.repo.GetLotes()
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
