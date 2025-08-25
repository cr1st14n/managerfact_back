package handlers

import (
	"managerfact/aplication/services"
	"managerfact/internal/domain/models"

	"github.com/gofiber/fiber/v2"
)

// TestConnectionConfig request para probar configuración
type TestConnectionConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

// TestConnectionByConfig prueba una configuración
func TestConnectionByConfig(c *fiber.Ctx) error {
	var req TestConnectionConfig
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Datos inválidos",
		})
	}

	config := &models.DbConnection{
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
	}

	result := services.TestConnection(config)

	return c.JSON(fiber.Map{
		"success": result.Success,
		"message": result.Message,
		"data":    result,
	})
}

// TestConnectionByID prueba una conexión existente
func TestConnectionByID(c *fiber.Ctx, connService services.DbConnectionService) error {
	// idStr := c.Params("id")
	// id, err := strconv.ParseUint(idStr, 10, 32)
	// if err != nil {
	// 	return c.Status(400).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "ID inválido",
	// 	})
	// }

	// result := connService.TestConnectionByID(uint(id))

	// return c.JSON(fiber.Map{
	// 	"success": result.Success,
	// 	"message": result.Message,
	// 	"data":    result,
	// })
	return nil
}
