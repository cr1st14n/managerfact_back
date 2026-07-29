package handlers

import (
	"errors"
	"managerfact/aplication/services"
	"managerfact/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	service *services.UsuarioService
}

func NewAuthHandler(s *services.UsuarioService) *AuthHandler {
	return &AuthHandler{service: s}
}

type loginRequest struct {
	CodigoUsuario string `json:"codigo_usuario"`
	Password      string `json:"password"`
}

// Login valida codigo_usuario + password y devuelve un JWT de sesión.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "error": err.Error()})
	}
	if req.CodigoUsuario == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "codigo_usuario y password son requeridos"})
	}

	usuario, err := h.service.Login(req.CodigoUsuario, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrCredencialesInvalidas) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error iniciando sesión", "error": err.Error()})
	}

	token, err := utils.GenerarTokenJWT(usuario.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error generando la sesión", "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Login exitoso",
		"data": fiber.Map{
			"token":   token,
			"usuario": usuario,
		},
	})
}

func (h *AuthHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/auth/login", h.Login)
}
