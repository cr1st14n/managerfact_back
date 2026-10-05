package middleware

import (
	"managerfact/aplication/services"
	"managerfact/pkg/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const UsuarioIDLocal = "usuario_id"

func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Token no proporcionado"})
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")

		claims, err := utils.ValidarTokenJWT(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Sesión inválida o expirada", "error": err.Error()})
		}

		c.Locals(UsuarioIDLocal, claims.UsuarioID)
		return c.Next()
	}
}

func RequireAdmin(usuarioService *services.UsuarioService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		usuarioID, ok := c.Locals(UsuarioIDLocal).(uint)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Sesión inválida"})
		}

		esAdmin, err := usuarioService.EsAdmin(usuarioID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error verificando permisos", "error": err.Error()})
		}
		if !esAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "No tienes permiso para acceder a este módulo"})
		}
		return c.Next()
	}
}

// requireNoConsultas debe limitar este rol a Reportes/DUAS y dependencias de solo lectura.
func RequireNoConsultas(usuarioService *services.UsuarioService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		usuarioID, ok := c.Locals(UsuarioIDLocal).(uint)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Sesión inválida"})
		}

		esConsultas, err := usuarioService.EsConsultas(usuarioID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error verificando permisos", "error": err.Error()})
		}
		if esConsultas {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "No tienes permiso para acceder a este módulo"})
		}
		return c.Next()
	}
}
