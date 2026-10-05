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

var errPingNoDisponible = errors.New("ping no disponible")

// El ping ICMP evita que guardar una conexión dependa de TLS o SQL Server.
func pingHost(host string) error {
	host = strings.TrimSpace(host)

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
