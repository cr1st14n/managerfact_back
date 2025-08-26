package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ValidarCampoRequerido valida campo requerido y retorna el string limpio
func ValidarCampoRequerido(errValidacion *[]string, campo, mensaje string) string {
	campoLimpio := strings.TrimSpace(campo)
	if campoLimpio == "" {
		*errValidacion = append(*errValidacion, mensaje)
		return ""
	}
	return campoLimpio
}

// ValidarEntero valida y convierte a entero
func ValidarEntero(errValidacion *[]string, campo, mensaje string) int {
	// campoLimpio := strings.TrimSpace(campo)
	campoLimpio := campo
	fmt.Println(campoLimpio)
	if campoLimpio == "" {
		*errValidacion = append(*errValidacion, mensaje)
		return 0
	}

	valor, err := strconv.Atoi(campoLimpio)
	if err != nil {
		*errValidacion = append(*errValidacion, err.Error())
		return 0
	}

	return valor
}

// ValidarEnteroOpcional valida y convierte a entero (campo opcional)
func ValidarEnteroOpcional(errValidacion *[]string, campo, mensaje string) int {
	campoLimpio := strings.TrimSpace(campo)
	if campoLimpio == "" {
		return 0 // Campo opcional, no agrega error
	}

	valor, err := strconv.Atoi(campoLimpio)
	if err != nil {
		*errValidacion = append(*errValidacion, mensaje)
		return 0
	}

	return valor
}

// ValidarFloat valida y convierte a float64
func ValidarFloat(errValidacion *[]string, campo, mensaje string) float64 {
	campoLimpio := strings.TrimSpace(campo)
	if campoLimpio == "" {
		*errValidacion = append(*errValidacion, mensaje)
		return 0.0
	}

	valor, err := strconv.ParseFloat(campoLimpio, 64)
	if err != nil {
		*errValidacion = append(*errValidacion, mensaje)
		return 0.0
	}

	return valor
}

// ValidarFloatOpcional valida y convierte a float64 (campo opcional)
func ValidarFloatOpcional(errValidacion *[]string, campo, mensaje string) float64 {
	campoLimpio := strings.TrimSpace(campo)
	if campoLimpio == "" {
		return 0.0 // Campo opcional, no agrega error
	}

	valor, err := strconv.ParseFloat(campoLimpio, 64)
	if err != nil {
		*errValidacion = append(*errValidacion, mensaje)
		return 0.0
	}

	return valor
}

// ValidarFecha valida y convierte a time.Time
func ValidarFecha(errValidacion *[]string, campo, mensaje string) time.Time {
	campoLimpio := strings.TrimSpace(campo)
	if campoLimpio == "" {
		*errValidacion = append(*errValidacion, mensaje)
		return time.Time{}
	}

	fecha, err := time.Parse("2006-01-02", campoLimpio)
	if err != nil {
		*errValidacion = append(*errValidacion, mensaje)
		return time.Time{}
	}

	return fecha
}

// ValidarFechaConFormato valida fecha con formato personalizado
func ValidarFechaConFormato(errValidacion *[]string, campo, formato, mensaje string) time.Time {
	campoLimpio := strings.TrimSpace(campo)
	if campoLimpio == "" {
		*errValidacion = append(*errValidacion, mensaje)
		return time.Time{}
	}

	fecha, err := time.Parse(formato, campoLimpio)
	if err != nil {
		*errValidacion = append(*errValidacion, mensaje)
		return time.Time{}
	}

	return fecha
}

// ValidarBooleano valida y convierte a bool
func ValidarBooleano(errValidacion *[]string, campo, mensaje string) bool {
	campoLimpio := strings.TrimSpace(strings.ToLower(campo))
	if campoLimpio == "" {
		*errValidacion = append(*errValidacion, mensaje)
		return false
	}

	switch campoLimpio {
	case "true", "1", "yes", "si", "sí", "verdadero", "on":
		return true
	case "false", "0", "no", "falso", "off":
		return false
	default:
		*errValidacion = append(*errValidacion, mensaje)
		return false
	}
}

// ValidarBooleanoOpcional valida booleano opcional (default false)
func ValidarBooleanoOpcional(errValidacion *[]string, campo, mensaje string) bool {
	campoLimpio := strings.TrimSpace(strings.ToLower(campo))
	if campoLimpio == "" {
		return false // Default para campo opcional
	}

	switch campoLimpio {
	case "true", "1", "yes", "si", "sí", "verdadero", "on":
		return true
	case "false", "0", "no", "falso", "off":
		return false
	default:
		*errValidacion = append(*errValidacion, mensaje)
		return false
	}
}

// ValidarCampoOpcional valida campo opcional y retorna string limpio
func ValidarCampoOpcional(errValidacion *[]string, campo string) string {
	return strings.TrimSpace(campo)
}

// ValidarRangoEntero valida entero dentro de un rango
func ValidarRangoEntero(errValidacion *[]string, campo, mensaje string, min, max int) int {
	valor := ValidarEntero(errValidacion, campo, mensaje)
	if valor < min || valor > max {
		*errValidacion = append(*errValidacion, mensaje+". Debe estar entre "+strconv.Itoa(min)+" y "+strconv.Itoa(max))
		return 0
	}
	return valor
}

// ValidarRangoFloat valida float dentro de un rango
func ValidarRangoFloat(errValidacion *[]string, campo, mensaje string, min, max float64) float64 {
	valor := ValidarFloat(errValidacion, campo, mensaje)
	if valor < min || valor > max {
		*errValidacion = append(*errValidacion, mensaje+". Debe estar entre "+strconv.FormatFloat(min, 'f', 2, 64)+" y "+strconv.FormatFloat(max, 'f', 2, 64))
		return 0.0
	}
	return valor
}
