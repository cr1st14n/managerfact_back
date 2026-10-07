package services

import "fmt"

type TipoFecha string

const (
	FechaEmision  TipoFecha = "emision"
	FechaEnvio    TipoFecha = "envio"
	FechaCreacion TipoFecha = "creacion"
)

// La columna SQL solo puede proceder de esta lista blanca.
func ColumnaTipoFecha(valor, defecto string) (string, error) {
	if valor == "" {
		valor = defecto
	}
	switch TipoFecha(valor) {
	case FechaEmision:
		return "sdf.fecha_emision", nil
	case FechaEnvio:
		return "sdf.fecha_envio", nil
	case FechaCreacion:
		return "sdf.created_date", nil
	default:
		return "", fmt.Errorf("tipoFecha inválido: use emision, envio o creacion")
	}
}
