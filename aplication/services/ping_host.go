package services

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// errPingNoDisponible: el comando ping no existe en el equipo donde corre el
// backend, así que no se pudo verificar (no significa que el host no responda).
var errPingNoDisponible = errors.New("ping no disponible")

// pingHost envía un solo ping ICMP al host. Solo comprueba que el equipo
// responda: no abre el puerto de SQL Server ni intenta iniciar sesión (eso
// lo hace el botón "Probar"). Evita que al guardar una conexión fallen cosas
// ajenas al host, como el certificado TLS del servidor.
func pingHost(host string) error {
	host = strings.TrimSpace(host)
	// El host viene del formulario del admin y se pasa como argumento de un
	// proceso: se rechaza todo lo que no sea un nombre o IP simple (opciones
	// como "-f" o metacaracteres).
	if host == "" || strings.HasPrefix(host, "-") || strings.ContainsAny(host, " \t\r\n;|&$`<>()\\'\"*?") {
		return fmt.Errorf("host inválido: %q", host)
	}

	var args []string
	switch runtime.GOOS {
	case "windows":
		args = []string{"-n", "1", "-w", "3000", host}
	case "darwin":
		args = []string{"-c", "1", "-W", "3000", host}
	default:
		args = []string{"-c", "1", "-W", "3", host}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "ping", args...).Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errPingNoDisponible
		}
		return fmt.Errorf("el servidor %s no responde al ping", host)
	}
	return nil
}
