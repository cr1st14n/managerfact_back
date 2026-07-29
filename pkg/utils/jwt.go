package utils

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ClaimsUsuario es el contenido del JWT de sesión: solo el ID del usuario,
// todo lo demás (accesos, acceso_total) se resuelve en cada request contra
// la BD, así un cambio de accesos aplica de inmediato sin esperar a que el
// token expire.
type ClaimsUsuario struct {
	UsuarioID uint `json:"usuario_id"`
	jwt.RegisteredClaims
}

func claveJWT() ([]byte, error) {
	secreto := os.Getenv("JWT_SECRET")
	if secreto == "" {
		return nil, fmt.Errorf("JWT_SECRET no está configurada")
	}
	return []byte(secreto), nil
}

func horasExpiracion() time.Duration {
	horas, err := strconv.Atoi(os.Getenv("JWT_EXPIRE_HOURS"))
	if err != nil || horas <= 0 {
		horas = 24
	}
	return time.Duration(horas) * time.Hour
}

// GenerarTokenJWT firma un token de sesión para el usuario dado.
func GenerarTokenJWT(usuarioID uint) (string, error) {
	clave, err := claveJWT()
	if err != nil {
		return "", err
	}

	claims := ClaimsUsuario{
		UsuarioID: usuarioID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(horasExpiracion())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	firmado, err := token.SignedString(clave)
	if err != nil {
		return "", fmt.Errorf("error firmando token: %w", err)
	}
	return firmado, nil
}

// ValidarTokenJWT valida la firma y expiración de un token y devuelve sus
// claims.
func ValidarTokenJWT(tokenString string) (*ClaimsUsuario, error) {
	clave, err := claveJWT()
	if err != nil {
		return nil, err
	}

	claims := &ClaimsUsuario{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return clave, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}
	return claims, nil
}
