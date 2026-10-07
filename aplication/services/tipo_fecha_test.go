package services

import "testing"

func TestColumnaTipoFechaListaBlanca(t *testing.T) {
	casos := []struct {
		valor, defecto, columna string
		valido                  bool
	}{
		{"emision", "creacion", "sdf.fecha_emision", true},
		{"envio", "emision", "sdf.fecha_envio", true},
		{"creacion", "emision", "sdf.created_date", true},
		{"", "creacion", "sdf.created_date", true},
		{"desconocido", "emision", "", false},
	}
	for _, caso := range casos {
		columna, err := ColumnaTipoFecha(caso.valor, caso.defecto)
		if (err == nil) != caso.valido || columna != caso.columna {
			t.Errorf("ColumnaTipoFecha(%q, %q) = (%q, %v)", caso.valor, caso.defecto, columna, err)
		}
	}
}
