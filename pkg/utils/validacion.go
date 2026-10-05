package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ValidarCampoRequerido(errValidacion *[]string, campo, mensaje string) string {
	campoLimpio := strings.TrimSpace(campo)
	if campoLimpio == "" {
		*errValidacion = append(*errValidacion, mensaje)
		return ""
	}
	return campoLimpio
}

func ValidarEntero(errValidacion *[]string, campo, mensaje string) int {

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

func ValidarEnteroOpcional(errValidacion *[]string, campo, mensaje string) int {
	campoLimpio := strings.TrimSpace(campo)
	if campoLimpio == "" {
		return 0 
	}

	valor, err := strconv.Atoi(campoLimpio)
	if err != nil {
		*errValidacion = append(*errValidacion, mensaje)
		return 0
	}

	return valor
}

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

func ValidarFloatOpcional(errValidacion *[]string, campo, mensaje string) float64 {
	campoLimpio := strings.TrimSpace(campo)
	if campoLimpio == "" {
		return 0.0 
	}

	valor, err := strconv.ParseFloat(campoLimpio, 64)
	if err != nil {
		*errValidacion = append(*errValidacion, mensaje)
		return 0.0
	}

	return valor
}

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

func ValidarBooleanoOpcional(errValidacion *[]string, campo, mensaje string) bool {
	campoLimpio := strings.TrimSpace(strings.ToLower(campo))
	if campoLimpio == "" {
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

func ValidarCampoOpcional(errValidacion *[]string, campo string) string {
	return strings.TrimSpace(campo)
}

func LimpiarListaOpcional(campos []string) []string {
	limpios := make([]string, 0, len(campos))
	for _, campo := range campos {
		campoLimpio := strings.TrimSpace(campo)
		if campoLimpio != "" {
			limpios = append(limpios, campoLimpio)
		}
	}
	return limpios
}

func ValidarRangoEntero(errValidacion *[]string, campo, mensaje string, min, max int) int {
	valor := ValidarEntero(errValidacion, campo, mensaje)
	if valor < min || valor > max {
		*errValidacion = append(*errValidacion, mensaje+". Debe estar entre "+strconv.Itoa(min)+" y "+strconv.Itoa(max))
		return 0
	}
	return valor
}

func ValidarRangoFloat(errValidacion *[]string, campo, mensaje string, min, max float64) float64 {
	valor := ValidarFloat(errValidacion, campo, mensaje)
	if valor < min || valor > max {
		*errValidacion = append(*errValidacion, mensaje+". Debe estar entre "+strconv.FormatFloat(min, 'f', 2, 64)+" y "+strconv.FormatFloat(max, 'f', 2, 64))
		return 0.0
	}
	return valor
}
