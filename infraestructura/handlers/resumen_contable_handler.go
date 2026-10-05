package handlers

import (
	"managerfact/aplication/services"
	"managerfact/infraestructura/middleware"
	"managerfact/internal/domain/models"
	"managerfact/pkg/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ResumenContableHandler struct {
	services.ResumenContableService
	usuarioService *services.UsuarioService
}

func NewResumenContableHandler(s *services.ResumenContableService, usuarioService *services.UsuarioService) *ResumenContableHandler {
	return &ResumenContableHandler{
		ResumenContableService: *s,
		usuarioService:         usuarioService,
	}
}

func (h *ResumenContableHandler) verificarAccesoTotal(c *fiber.Ctx) bool {
	usuarioID, ok := c.Locals(middleware.UsuarioIDLocal).(uint)
	if !ok {
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Sesión inválida"})
		return false
	}
	tieneAccesoTotal, err := h.usuarioService.TieneAccesoTotal(usuarioID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error verificando accesos",
			"error":   err.Error(),
		})
		return false
	}
	if !tieneAccesoTotal {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Este reporte junta todas las sucursales de la empresa: requiere acceso total",
		})
		return false
	}
	return true
}

func (h *ResumenContableHandler) leerFechas(c *fiber.Ctx) (fechaDesde, fechaHastaExclusiva time.Time, ok bool) {
	var errValidacion []string
	fechaDesde = utils.ValidarFecha(&errValidacion, c.Query("fechaDesde"), "El campo fechaDesde es requerido")
	fechaHasta := utils.ValidarFecha(&errValidacion, c.Query("fechaHasta"), "El campo fechaHasta es requerido")
	if len(errValidacion) > 0 {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Datos inválidos",
			"errors":  errValidacion,
		})
		return time.Time{}, time.Time{}, false
	}
	if fechaHasta.Before(fechaDesde) {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "fechaHasta no puede ser anterior a fechaDesde"})
		return time.Time{}, time.Time{}, false
	}

	return fechaDesde, fechaHasta.AddDate(0, 0, 1), true
}

func (h *ResumenContableHandler) leerPeriodo(c *fiber.Ctx) (idServer int64, fechaDesde, fechaHastaExclusiva time.Time, ok bool) {
	idServer, errServer := strconv.ParseInt(c.Query("idServer"), 10, 64)
	if errServer != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "El campo idServer es requerido"})
		return 0, time.Time{}, time.Time{}, false
	}

	fechaDesde, fechaHastaExclusiva, ok = h.leerFechas(c)
	return idServer, fechaDesde, fechaHastaExclusiva, ok
}

// Usar el agregado SQL en pantalla para no traer miles de facturas solo para sumarlas.
func (h *ResumenContableHandler) LibroVentasIvaResumen(c *fiber.Ctx) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	idServer, fechaDesde, fechaHastaExclusiva, ok := h.leerPeriodo(c)
	if !ok {
		return nil
	}

	data, err := h.ResumenContableService.LibroVentasIvaResumen(idServer, fechaDesde, fechaHastaExclusiva)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Resumen del libro de ventas IVA",
		"data":    data,
	})
}

func (h *ResumenContableHandler) LibroVentasIva(c *fiber.Ctx) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	idServer, fechaDesde, fechaHastaExclusiva, ok := h.leerPeriodo(c)
	if !ok {
		return nil
	}

	data, err := h.ResumenContableService.LibroVentasIva(idServer, fechaDesde, fechaHastaExclusiva)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Libro de ventas IVA",
		"data":    data,
	})
}

func (h *ResumenContableHandler) ResumenMensualDebitoFiscal(c *fiber.Ctx) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	idServer, fechaDesde, fechaHastaExclusiva, ok := h.leerPeriodo(c)
	if !ok {
		return nil
	}

	data, err := h.ResumenContableService.ResumenMensualDebitoFiscal(idServer, fechaDesde, fechaHastaExclusiva)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Resumen mensual de débito fiscal",
		"data":    data,
	})
}

func (h *ResumenContableHandler) totalesTodosServidores(
	c *fiber.Ctx,
	mensaje string,
	consultar func(ambiente string, desde, hasta time.Time) ([]models.ServidorTotales, error),
) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	fechaDesde, fechaHastaExclusiva, ok := h.leerFechas(c)
	if !ok {
		return nil
	}
	ambiente, ambienteValido := models.NormalizarAmbiente(c.Query("ambiente"))
	if !ambienteValido {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Ambiente inválido"})
	}

	data, err := consultar(ambiente, fechaDesde, fechaHastaExclusiva)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": mensaje,
		"data":    data,
	})
}

func (h *ResumenContableHandler) FacturasAnuladas(c *fiber.Ctx) error {
	return h.totalesTodosServidores(c, "Facturas anuladas del período", h.ResumenContableService.FacturasAnuladas)
}

func (h *ResumenContableHandler) FacturasValidas(c *fiber.Ctx) error {
	return h.totalesTodosServidores(c, "Facturas válidas del período", h.ResumenContableService.FacturasValidas)
}

func (h *ResumenContableHandler) NotasCreditoDebito(c *fiber.Ctx) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	idServer, fechaDesde, fechaHastaExclusiva, ok := h.leerPeriodo(c)
	if !ok {
		return nil
	}

	data, err := h.ResumenContableService.NotasCreditoDebito(idServer, fechaDesde, fechaHastaExclusiva)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Notas de crédito/débito y conciliación",
		"data":    data,
	})
}

func (h *ResumenContableHandler) FacturacionPorActividad(c *fiber.Ctx) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	idServer, fechaDesde, fechaHastaExclusiva, ok := h.leerPeriodo(c)
	if !ok {
		return nil
	}

	data, err := h.ResumenContableService.FacturacionPorActividad(idServer, fechaDesde, fechaHastaExclusiva)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Facturación por actividad económica",
		"data":    data,
	})
}

func (h *ResumenContableHandler) IngresosPorProducto(c *fiber.Ctx) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	idServer, fechaDesde, fechaHastaExclusiva, ok := h.leerPeriodo(c)
	if !ok {
		return nil
	}

	data, err := h.ResumenContableService.IngresosPorProducto(idServer, fechaDesde, fechaHastaExclusiva)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Ingresos por producto",
		"data":    data,
	})
}

func (h *ResumenContableHandler) IngresosPorSucursalPos(c *fiber.Ctx) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	idServer, fechaDesde, fechaHastaExclusiva, ok := h.leerPeriodo(c)
	if !ok {
		return nil
	}

	data, err := h.ResumenContableService.IngresosPorSucursalPos(idServer, fechaDesde, fechaHastaExclusiva)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Ingresos por sucursal y punto de venta",
		"data":    data,
	})
}

func (h *ResumenContableHandler) IngresosPorMetodoPago(c *fiber.Ctx) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	idServer, fechaDesde, fechaHastaExclusiva, ok := h.leerPeriodo(c)
	if !ok {
		return nil
	}

	data, err := h.ResumenContableService.IngresosPorMetodoPago(idServer, fechaDesde, fechaHastaExclusiva)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Ingresos por método de pago",
		"data":    data,
	})
}

func (h *ResumenContableHandler) IngresosPorMoneda(c *fiber.Ctx) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	idServer, fechaDesde, fechaHastaExclusiva, ok := h.leerPeriodo(c)
	if !ok {
		return nil
	}

	data, err := h.ResumenContableService.IngresosPorMoneda(idServer, fechaDesde, fechaHastaExclusiva)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Ingresos por moneda",
		"data":    data,
	})
}

// Mantener el límite TOP: el número de clientes distintos puede ser grande.
func (h *ResumenContableHandler) FacturacionPorCliente(c *fiber.Ctx) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	idServer, fechaDesde, fechaHastaExclusiva, ok := h.leerPeriodo(c)
	if !ok {
		return nil
	}

	top := 50
	if topStr := c.Query("top"); topStr != "" {
		var errValidacion []string
		topValor := utils.ValidarEnteroOpcional(&errValidacion, topStr, "El campo top debe ser numérico")
		if len(errValidacion) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Datos inválidos",
				"errors":  errValidacion,
			})
		}
		if topValor < 1 || topValor > 500 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "El campo top debe estar entre 1 y 500"})
		}
		top = topValor
	}

	data, err := h.ResumenContableService.FacturacionPorCliente(idServer, fechaDesde, fechaHastaExclusiva, top)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Facturación por cliente",
		"data":    data,
	})
}

func (h *ResumenContableHandler) EmisionPorUsuario(c *fiber.Ctx) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	idServer, fechaDesde, fechaHastaExclusiva, ok := h.leerPeriodo(c)
	if !ok {
		return nil
	}

	data, err := h.ResumenContableService.EmisionPorUsuario(idServer, fechaDesde, fechaHastaExclusiva)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Emisión por usuario",
		"data":    data,
	})
}

func (h *ResumenContableHandler) RegisterRoutes(router fiber.Router) {
	resumen := router.Group("/resumen-contable")
	resumen.Get("/libro-ventas-iva", h.LibroVentasIva)
	resumen.Get("/libro-ventas-iva/resumen", h.LibroVentasIvaResumen)
	resumen.Get("/resumen-mensual-debito-fiscal", h.ResumenMensualDebitoFiscal)
	resumen.Get("/facturas-anuladas", h.FacturasAnuladas)
	resumen.Get("/facturas-validas", h.FacturasValidas)
	resumen.Get("/notas-credito-debito", h.NotasCreditoDebito)
	resumen.Get("/facturacion-por-actividad", h.FacturacionPorActividad)
	resumen.Get("/ingresos-por-producto", h.IngresosPorProducto)
	resumen.Get("/ingresos-por-sucursal-pos", h.IngresosPorSucursalPos)
	resumen.Get("/ingresos-por-metodo-pago", h.IngresosPorMetodoPago)
	resumen.Get("/ingresos-por-moneda", h.IngresosPorMoneda)
	resumen.Get("/facturacion-por-cliente", h.FacturacionPorCliente)
	resumen.Get("/emision-por-usuario", h.EmisionPorUsuario)
}
