package handlers

import (
	"fmt"
	"managerfact/aplication/services"
	"managerfact/infraestructura/middleware"
	"managerfact/internal/domain/models"
	"managerfact/pkg/utils"
	"strconv"
	"strings"

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
// Devuelve true si el acceso es válido. Si es false, ya escribió la
// respuesta de error en c y el caller debe cortar sin llamar a c.JSON de
// nuevo (y sin devolver un error a Fiber, que lo pisaría con un 500 vía el
// ErrorHandler global).
func (h *ConsultasHandler) verificarAccesoSucursal(c *fiber.Ctx, codigoSucursalSin string) bool {
	usuarioID, ok := c.Locals(middleware.UsuarioIDLocal).(uint)
	if !ok {
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Sesión inválida"})
		return false
	}

	if codigoSucursalSin == "" {
		tieneAccesoTotal, err := h.usuarioService.TieneAccesoTotal(usuarioID)
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error verificando accesos", "error": err.Error()})
			return false
		}
		if !tieneAccesoTotal {
			fmt.Println("Usuario", usuarioID, " NO tiene acceso total y no indicó sucursal")
			c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Debes indicar una sucursal a la que tengas acceso"})
			return false
		}

		return true
	}

	codigo, err := strconv.Atoi(codigoSucursalSin)
	fmt.Println("Usuario", usuarioID, "SI tiene acceso a la sucursal", codigoSucursalSin)

	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "codigoSucursalSin inválido"})
		return false
	}
	permitido, err := h.usuarioService.TieneAccesoSucursal(usuarioID, codigo)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error verificando accesos", "error": err.Error()})
		return false
	}
	if !permitido {
		fmt.Println("Usuario", usuarioID, "NO tiene acceso a la sucursal", codigoSucursalSin)
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "No tiene acceso a la sucursal " + codigoSucursalSin})
		return false
	}

	return true
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
	fmt.Printf("sucursal requerida %s", dataIn.CodigoSucursalSin)
	if !h.verificarAccesoSucursal(c, dataIn.CodigoSucursalSin) {
		return nil
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

// FacturasMes devuelve las facturas verificadas de un mes para una sucursal y
// un código de producto (base de la descarga mensual a Excel).
func (h *ConsultasHandler) FacturasMes(c *fiber.Ctx) error {
	idServer, errSrv := strconv.ParseInt(c.Query("idServer"), 10, 64)
	idSucursal, errSuc := strconv.Atoi(c.Query("sucursal"))
	codigoSin := c.Query("codigoSucursalSin")
	codigoSinInt, errSin := strconv.Atoi(codigoSin)
	producto := strings.TrimSpace(c.Query("codigoProducto"))
	anio, errAnio := strconv.Atoi(c.Query("anio"))
	mes, errMes := strconv.Atoi(c.Query("mes"))

	if errSrv != nil || errSuc != nil || errSin != nil || producto == "" ||
		errAnio != nil || errMes != nil || mes < 1 || mes > 12 || anio < 2000 || anio > 2100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Parámetros requeridos: idServer, sucursal, codigoSucursalSin, codigoProducto, anio, mes",
		})
	}

	if !h.verificarAccesoSucursal(c, codigoSin) {
		return nil
	}

	data, err := h.ConsultasService.FacturasMes(idServer, idSucursal, codigoSinInt, producto, anio, mes)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "Facturas del mes",
		"data":    data,
	})
}

func (h *ConsultasHandler) RegisterRoutes(router fiber.Router) {
	connections := router.Group("/consultar")

	connections.Post("/", h.DataFacturas)
	connections.Get("/sucursales", h.Sucursales)
	connections.Get("/duas", h.BuscarDuas)
	connections.Get("/facturas-mes", h.FacturasMes)
}
