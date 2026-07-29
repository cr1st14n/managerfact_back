package handlers

import (
	"managerfact/aplication/services"
	"managerfact/internal/domain/models"
	"managerfact/pkg/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type UsuarioHandler struct {
	service *services.UsuarioService
}

func NewUsuarioHandler(s *services.UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{service: s}
}

type usuarioRequest struct {
	Nombre        string `json:"nombre"`
	CI            string `json:"ci"`
	Cargo         string `json:"cargo"`
	CodigoUsuario string `json:"codigo_usuario"`
	Rol           string `json:"rol"`
	RegionalID    *uint  `json:"regional_id"`
	SucursalID    *uint  `json:"sucursal_id"`
	IsActive      *bool  `json:"is_active,omitempty"`
}

func (h *UsuarioHandler) validar(req *usuarioRequest) []string {
	var errValidacion []string
	req.Nombre = utils.ValidarCampoRequerido(&errValidacion, req.Nombre, "El campo nombre es requerido")
	req.CI = utils.ValidarCampoRequerido(&errValidacion, req.CI, "El campo ci es requerido")
	req.CodigoUsuario = utils.ValidarCampoRequerido(&errValidacion, req.CodigoUsuario, "El campo codigo_usuario es requerido")
	req.Cargo = utils.ValidarCampoOpcional(&errValidacion, req.Cargo)
	if req.Rol != models.RolAdmin && req.Rol != models.RolOperador {
		errValidacion = append(errValidacion, "El campo rol debe ser 'admin' u 'operador'")
	}
	return errValidacion
}

func (h *UsuarioHandler) Create(c *fiber.Ctx) error {
	var req usuarioRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "error": err.Error()})
	}

	if errValidacion := h.validar(&req); len(errValidacion) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "errors": errValidacion})
	}

	usuario, err := h.service.Crear(services.CrearUsuarioInput{
		Nombre:        req.Nombre,
		CI:            req.CI,
		Cargo:         req.Cargo,
		CodigoUsuario: req.CodigoUsuario,
		Rol:           req.Rol,
		RegionalID:    req.RegionalID,
		SucursalID:    req.SucursalID,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error creando usuario", "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Usuario creado exitosamente (contraseña inicial = CI)",
		"data":    usuario,
	})
}

func (h *UsuarioHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}

	var req usuarioRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "error": err.Error()})
	}

	if errValidacion := h.validar(&req); len(errValidacion) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "errors": errValidacion})
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	usuario, err := h.service.Actualizar(services.ActualizarUsuarioInput{
		ID:            uint(id),
		Nombre:        req.Nombre,
		CI:            req.CI,
		Cargo:         req.Cargo,
		CodigoUsuario: req.CodigoUsuario,
		Rol:           req.Rol,
		RegionalID:    req.RegionalID,
		SucursalID:    req.SucursalID,
		IsActive:      isActive,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error actualizando usuario", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Usuario actualizado exitosamente", "data": usuario})
}

func (h *UsuarioHandler) GetAll(c *fiber.Ctx) error {
	usuarios, err := h.service.ListarTodos()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error obteniendo usuarios", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Usuarios obtenidos exitosamente", "data": usuarios})
}

func (h *UsuarioHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}
	usuario, err := h.service.ObtenerPorID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Usuario no encontrado", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Usuario encontrado", "data": usuario})
}

func (h *UsuarioHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}
	if err := h.service.Eliminar(uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error eliminando usuario", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Usuario eliminado exitosamente"})
}

func (h *UsuarioHandler) ResetPassword(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}
	if err := h.service.ResetPassword(uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error reseteando contraseña", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Contraseña restablecida al CI del usuario"})
}

type accesosRequest struct {
	AccesoTotal   bool   `json:"acceso_total"`
	RegionalesIDs []uint `json:"regionales_ids"`
	SucursalesIDs []uint `json:"sucursales_ids"`
}

func (h *UsuarioHandler) SetAccesos(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}

	var req accesosRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Datos inválidos", "error": err.Error()})
	}

	if err := h.service.ConfigurarAccesos(uint(id), services.AccesosInput{
		AccesoTotal:   req.AccesoTotal,
		RegionalesIDs: req.RegionalesIDs,
		SucursalesIDs: req.SucursalesIDs,
	}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error configurando accesos", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Accesos actualizados exitosamente"})
}

func (h *UsuarioHandler) GetAccesos(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}
	accesos, err := h.service.ObtenerAccesos(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error obteniendo accesos", "error": err.Error()})
	}

	// Mismo shape que antes ([{regional_id}], [{sucursal_id}]) para no tener
	// que tocar el front, que ya sabe leer estos dos campos.
	regionales := make([]fiber.Map, len(accesos.RegionalesIDs))
	for i, id := range accesos.RegionalesIDs {
		regionales[i] = fiber.Map{"regional_id": id}
	}
	sucursales := make([]fiber.Map, len(accesos.SucursalesIDs))
	for i, id := range accesos.SucursalesIDs {
		sucursales[i] = fiber.Map{"sucursal_id": id}
	}

	return c.JSON(fiber.Map{
		"message": "Accesos obtenidos exitosamente",
		"data": fiber.Map{
			"regionales": regionales,
			"sucursales": sucursales,
		},
	})
}

// GetSucursalesPermitidas resuelve el catálogo efectivo de sucursales a las
// que el usuario tiene acceso. Todavía no está conectado a ningún select del
// front (el login sigue pendiente); queda listo para cuando exista sesión.
func (h *UsuarioHandler) GetSucursalesPermitidas(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "ID inválido"})
	}
	sucursales, err := h.service.SucursalesPermitidas(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error resolviendo sucursales permitidas", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Sucursales permitidas obtenidas exitosamente", "data": sucursales})
}

func (h *UsuarioHandler) GetRegionales(c *fiber.Ctx) error {
	regionales, err := h.service.ListarRegionales()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error obteniendo regionales", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Regionales obtenidas exitosamente", "data": regionales})
}

func (h *UsuarioHandler) GetSucursalesCatalogo(c *fiber.Ctx) error {
	sucursales, err := h.service.ListarSucursalesCatalogo()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error obteniendo sucursales", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Sucursales obtenidas exitosamente", "data": sucursales})
}

// RegisterRoutes registra las rutas de usuarios y los catálogos de
// regionales/sucursales que usa el formulario de accesos, todas detrás de
// requireAdmin (los operadores no pueden ver ni operar este módulo). Los
// prefijos "/usuarios", "/regionales" y "/sucursales-catalogo" scopean el
// middleware solo a estas rutas, no a las demás del grupo protegido.
func (h *UsuarioHandler) RegisterRoutes(router fiber.Router, requireAdmin fiber.Handler) {
	usuarios := router.Group("/usuarios", requireAdmin)
	usuarios.Post("/", h.Create)
	usuarios.Get("/", h.GetAll)
	usuarios.Get("/:id", h.GetByID)
	usuarios.Put("/:id", h.Update)
	usuarios.Delete("/:id", h.Delete)
	usuarios.Post("/:id/reset-password", h.ResetPassword)
	usuarios.Get("/:id/accesos", h.GetAccesos)
	usuarios.Put("/:id/accesos", h.SetAccesos)
	usuarios.Get("/:id/sucursales-permitidas", h.GetSucursalesPermitidas)

	router.Get("/regionales", requireAdmin, h.GetRegionales)
	router.Get("/sucursales-catalogo", requireAdmin, h.GetSucursalesCatalogo)
}
