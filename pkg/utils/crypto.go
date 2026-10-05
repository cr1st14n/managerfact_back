package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

func claveCifrado() ([]byte, error) {
	texto := strings.TrimSpace(os.Getenv("FACTURADOR_TOKEN_KEY"))
	if texto == "" {
		return nil, fmt.Errorf("FACTURADOR_TOKEN_KEY no está configurada")
	}
	clave := sha256.Sum256([]byte(texto))
	return clave[:], nil
}

func Encrypt(texto string) (string, error) {
	clave, err := claveCifrado()
	if err != nil {
		return "", err
	}

	bloque, err := aes.NewCipher(clave)
	if err != nil {
		return "", fmt.Errorf("error inicializando cifrado: %w", err)
	}

	gcm, err := cipher.NewGCM(bloque)
	if err != nil {
		return "", fmt.Errorf("error inicializando GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("error generando nonce: %w", err)
	}

	cifrado := gcm.Seal(nonce, nonce, []byte(texto), nil)
	return base64.StdEncoding.EncodeToString(cifrado), nil
}

func Decrypt(textoCifrado string) (string, error) {
	clave, err := claveCifrado()
	if err != nil {
		return "", err
	}

	datos, err := base64.StdEncoding.DecodeString(textoCifrado)
	if err != nil {
		return "", fmt.Errorf("valor cifrado inválido (base64): %w", err)
	}

	bloque, err := aes.NewCipher(clave)
	if err != nil {
		return "", fmt.Errorf("error inicializando cifrado: %w", err)
	}

	gcm, err := cipher.NewGCM(bloque)
	if err != nil {
		return "", fmt.Errorf("error inicializando GCM: %w", err)
	}

	tamanoNonce := gcm.NonceSize()
	if len(datos) < tamanoNonce {
		return "", fmt.Errorf("valor cifrado corrupto")
	}

	nonce, ciphertext := datos[:tamanoNonce], datos[tamanoNonce:]
	texto, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("error descifrando: %w", err)
	}

	return string(texto), nil
}
