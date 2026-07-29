package middleware

import (
	"managerfact/aplication/services"
	"managerfact/pkg/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// UsuarioIDLocal es la key usada para guardar el usuario_id del token en
// c.Locals una vez validado por RequireAuth.
const UsuarioIDLocal = "usuario_id"

// RequireAuth exige un JWT válido en el header Authorization ("Bearer
// <token>") y deja el usuario_id disponible en c.Locals para los handlers
// siguientes.
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

// RequireAdmin exige que el usuario autenticado (usuario_id ya puesto en
// Locals por RequireAuth, que debe ir encadenado antes) tenga rol "admin".
// Usado para bloquear los módulos de Usuarios y Sucursales Facturador a los
// operadores.
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
