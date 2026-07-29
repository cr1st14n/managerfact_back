package handlers

import (
	"managerfact/aplication/services"
	"managerfact/infraestructura/middleware"
	"managerfact/internal/domain/models"
	"managerfact/pkg/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ConsultasHandler struct {
	services.ConsultasService
	usuarioService *services.UsuarioService
}

func NewConsultasHandler(s *services.ConsultasService, usuarioService *services.UsuarioService) *ConsultasHandler {
	return &ConsultasHandler{
		ConsultasService: *s,
		usuarioService:   usuarioService,
	}
}

// verificarAccesoSucursal exige que el usuario autenticado (usuario_id
// puesto en Locals por middleware.RequireAuth) tenga permiso sobre el
// codigoSucursalSin solicitado. Si viene vacío, solo se permite a usuarios
// con acceso total — no se puede pedir "todas las sucursales" sin tenerlo.
// Devuelve nil si el acceso es válido, o la respuesta de error ya escrita.
func (h *ConsultasHandler) verificarAccesoSucursal(c *fiber.Ctx, codigoSucursalSin string) error {
	usuarioID, ok := c.Locals(middleware.UsuarioIDLocal).(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Sesión inválida"})
	}

	if codigoSucursalSin == "" {
		tieneAccesoTotal, err := h.usuarioService.TieneAccesoTotal(usuarioID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error verificando accesos", "error": err.Error()})
		}
		if !tieneAccesoTotal {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Debes indicar una sucursal a la que tengas acceso"})
		}
		return nil
	}

	codigo, err := strconv.Atoi(codigoSucursalSin)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "codigoSucursalSin inválido"})
	}
	permitido, err := h.usuarioService.TieneAccesoSucursal(usuarioID, codigo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error verificando accesos", "error": err.Error()})
	}
	if !permitido {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "No tiene acceso a la sucursal " + codigoSucursalSin})
	}
	return nil
}

func (h *ConsultasHandler) DataFacturas(c *fiber.Ctx) error {
	var dataIn models.Json_consulta_data

	if err := c.BodyParser(&dataIn); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Datos inválidos",
			"error":   err.Error(),
		})
	}

	var errValidacion []string
	dataIn.IdFacturador = utils.ValidarCampoRequerido(&errValidacion, dataIn.IdFacturador, "El campo idServer es requerido")
	if dataIn.NumeroFactura != "" {
		utils.ValidarEntero(&errValidacion, dataIn.NumeroFactura, "El campo numeroFactura es requerido")
	}
	dataIn.CodigoProducto = utils.LimpiarListaOpcional(dataIn.CodigoProducto)
	FechaDesde_parse := utils.ValidarFecha(&errValidacion, dataIn.FechaDesde, "El campo fechaDesde es requerido")
	FechaHasta_parse := utils.ValidarFecha(&errValidacion, dataIn.FechaHasta, "El campo fechaHasta es requerido")
	dataIn.Sucursal = utils.ValidarCampoOpcional(&errValidacion, dataIn.Sucursal)
	dataIn.FechaDesde = FechaDesde_parse.Format("2006-01-02")
	dataIn.FechaHasta = FechaHasta_parse.Format("2006-01-02")

	dataIn.CodigoIntegracion = utils.ValidarCampoOpcional(&errValidacion, dataIn.CodigoIntegracion)
	dataIn.CodigoCliente = utils.ValidarCampoOpcional(&errValidacion, dataIn.CodigoCliente)
	dataIn.CUF = utils.ValidarCampoOpcional(&errValidacion, dataIn.CUF)
	dataIn.EstadoDocumentoFiscal = utils.ValidarCampoOpcional(&errValidacion, dataIn.EstadoDocumentoFiscal)
	if dataIn.CodigoSucursalSin != "" {
		utils.ValidarEnteroOpcional(&errValidacion, dataIn.CodigoSucursalSin, "El campo codigoSucursalSin debe ser numérico")
	}
	if dataIn.TipoEmision != "" {
		utils.ValidarEnteroOpcional(&errValidacion, dataIn.TipoEmision, "El campo tipoEmision debe ser numérico")
	}

	if len(errValidacion) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Datos inválidos",
			"errors":  errValidacion,
		})
	}

	if err := h.verificarAccesoSucursal(c, dataIn.CodigoSucursalSin); err != nil {
		return err
	}

	data, errDaS := h.ConsultasService.DataFacturas(dataIn)
	if errDaS != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   errDaS.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Informacion de factura",
		"data":    data,
	})
}

func (h *ConsultasHandler) Sucursales(c *fiber.Ctx) error {
	var idServer = c.Query("idServer")
	data, err := h.ConsultasService.Sucursales(idServer)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Sucursales registradas",
		"data":    data,
	})
	// return nil
}

func (h *ConsultasHandler) BuscarDuas(c *fiber.Ctx) error {
	idServer := c.Query("idServer")
	if idServer == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "El campo idServer es requerido",
		})
	}

	params := models.DuasBusquedaParams{
		Nombre:      c.Query("nombre"),
		Apellido:    c.Query("apellido"),
		NumeroVuelo: c.Query("numero_vuelo"),
		FechaDesde:  c.Query("fecha_desde"),
		FechaHasta:  c.Query("fecha_hasta"),
		Asiento:     c.Query("asiento"),
		Ticket:      c.Query("ticket"),
	}

	data, err := h.ConsultasService.BuscarDuas(idServer, params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error en consulta DUAS",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Resultados DUAS obtenidos",
		"data":    data,
	})
}

func (h *ConsultasHandler) RegisterRoutes(router fiber.Router) {
	connections := router.Group("/consultar")

	connections.Post("/", h.DataFacturas)
	connections.Get("/sucursales", h.Sucursales)
	connections.Get("/duas", h.BuscarDuas)
}
