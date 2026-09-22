package handlers

import (
	"managerfact/aplication/services"
	"managerfact/infraestructura/middleware"
	"managerfact/pkg/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ResumenContableHandler expone los reportes del módulo Resumen Contable
// (Libro Ventas IVA es el primero; ver ClicReportes.md para los que siguen).
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

// verificarAccesoTotal exige que el usuario autenticado tenga acceso
// nacional. Los reportes de este módulo juntan TODAS las sucursales de la
// empresa (a diferencia de /consultar, no admiten filtrar por una sola), así
// que no pueden reusar verificarAccesoSucursal de ConsultasHandler: acá es
// acceso total sin excepción, para no mostrarle a un usuario con acceso
// limitado sucursales que no le corresponden. Igual que su equivalente en
// ConsultasHandler, si devuelve false ya escribió la respuesta de error.
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

// leerPeriodo valida idServer/fechaDesde/fechaHasta (query params, comunes a
// todos los reportes de este handler) y devuelve fechaHasta ya como límite
// exclusivo (inicio del día siguiente), para no perder facturas del último
// día por la hora. Si ok es false ya escribió la respuesta de error.
func (h *ResumenContableHandler) leerPeriodo(c *fiber.Ctx) (idServer int64, fechaDesde, fechaHastaExclusiva time.Time, ok bool) {
	idServer, errServer := strconv.ParseInt(c.Query("idServer"), 10, 64)
	if errServer != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "El campo idServer es requerido"})
		return 0, time.Time{}, time.Time{}, false
	}

	var errValidacion []string
	fechaDesde = utils.ValidarFecha(&errValidacion, c.Query("fechaDesde"), "El campo fechaDesde es requerido")
	fechaHasta := utils.ValidarFecha(&errValidacion, c.Query("fechaHasta"), "El campo fechaHasta es requerido")
	if len(errValidacion) > 0 {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Datos inválidos",
			"errors":  errValidacion,
		})
		return 0, time.Time{}, time.Time{}, false
	}
	if fechaHasta.Before(fechaDesde) {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "fechaHasta no puede ser anterior a fechaDesde"})
		return 0, time.Time{}, time.Time{}, false
	}

	return idServer, fechaDesde, fechaHasta.AddDate(0, 0, 1), true
}

// LibroVentasIvaResumen es el total de control por sucursal: lo que se
// muestra en pantalla. Liviano a propósito (agregado en SQL, no factura por
// factura) para no mover miles de filas solo para sumarlas.
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

// LibroVentasIva es el detalle factura por factura, pensado solo para la
// descarga a Excel: el front no lo pinta en una tabla (ver
// LibroVentasIvaResumen para lo que sí se muestra en pantalla).
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

// ResumenMensualDebitoFiscal es el segundo reporte del módulo: el total de
// control por mes de TODA la empresa (sin desglosar por sucursal), para
// revisar varios meses de un vistazo y cuadrar contra lo declarado al SIN.
// Ver ClicReportes.md sección 2.
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

// FacturasAnuladas es el tercer reporte del módulo: el detalle de
// solicitudes de anulación del período (filtradas por fecha de envío de la
// anulación, no de emisión de la factura). Ver ClicReportes.md sección 3.
func (h *ResumenContableHandler) FacturasAnuladas(c *fiber.Ctx) error {
	if !h.verificarAccesoTotal(c) {
		return nil
	}
	idServer, fechaDesde, fechaHastaExclusiva, ok := h.leerPeriodo(c)
	if !ok {
		return nil
	}

	data, err := h.ResumenContableService.FacturasAnuladas(idServer, fechaDesde, fechaHastaExclusiva)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error de consulta",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Facturas anuladas del período",
		"data":    data,
	})
}

// NotasCreditoDebito es el cuarto reporte del módulo: notas de
// crédito/débito y de conciliación del período. Ver ClicReportes.md
// sección 4.
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

// FacturacionPorActividad es el quinto reporte del módulo: total facturado
// y base del débito fiscal por actividad económica de cabecera. Ver
// ClicReportes.md sección 5.
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

// IngresosPorProducto es el sexto reporte del módulo: total por producto
// (mapea a cuentas de ingreso). Ver ClicReportes.md sección 6.
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

// IngresosPorSucursalPos es el séptimo reporte del módulo: total por
// sucursal y punto de venta. Ver ClicReportes.md sección 7.
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

// IngresosPorMetodoPago es el octavo reporte del módulo: total por método
// de pago. Ver ClicReportes.md sección 8.
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

// IngresosPorMoneda es el noveno reporte del módulo: total por moneda y
// tipo de cambio. Ver ClicReportes.md sección 9.
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

// FacturacionPorCliente es el décimo reporte del módulo: los "top" clientes
// por total facturado (default 50, ver ClicReportes.md sección 10 y el
// comentario de facturacionPorClienteQuery). El límite es a propósito: el
// número de clientes distintos puede ser grande.
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

// EmisionPorUsuario es el onceavo reporte del módulo: arqueo / control
// interno por usuario, sucursal y día. Ver ClicReportes.md sección 11.
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
	resumen.Get("/notas-credito-debito", h.NotasCreditoDebito)
	resumen.Get("/facturacion-por-actividad", h.FacturacionPorActividad)
	resumen.Get("/ingresos-por-producto", h.IngresosPorProducto)
	resumen.Get("/ingresos-por-sucursal-pos", h.IngresosPorSucursalPos)
	resumen.Get("/ingresos-por-metodo-pago", h.IngresosPorMetodoPago)
	resumen.Get("/ingresos-por-moneda", h.IngresosPorMoneda)
	resumen.Get("/facturacion-por-cliente", h.FacturacionPorCliente)
	resumen.Get("/emision-por-usuario", h.EmisionPorUsuario)
}
