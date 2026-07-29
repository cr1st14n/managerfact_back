package services

import (
	"errors"
	"fmt"
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"

	"golang.org/x/crypto/bcrypt"
)

type UsuarioService struct {
	repo *repositories.UsuarioRepository
}

func NewUsuarioService(r *repositories.UsuarioRepository) *UsuarioService {
	return &UsuarioService{repo: r}
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error generando contraseña: %w", err)
	}
	return string(hash), nil
}

type CrearUsuarioInput struct {
	Nombre        string
	CI            string
	Cargo         string
	CodigoUsuario string
	RegionalID    *uint
	SucursalID    *uint
}

// Crear registra un usuario nuevo. La contraseña inicial es siempre su CI
// (hasheada), tal como se pidió: "pass por defecto el CI".
func (s *UsuarioService) Crear(input CrearUsuarioInput) (*models.Usuario, error) {
	passwordHash, err := hashPassword(input.CI)
	if err != nil {
		return nil, err
	}

	usuario := &models.Usuario{
		Nombre:        input.Nombre,
		CI:            input.CI,
		Cargo:         input.Cargo,
		CodigoUsuario: input.CodigoUsuario,
		RegionalID:    input.RegionalID,
		SucursalID:    input.SucursalID,
		PasswordHash:  passwordHash,
		IsActive:      true,
	}
	if err := s.repo.Create(usuario); err != nil {
		return nil, err
	}
	return usuario, nil
}

type ActualizarUsuarioInput struct {
	ID            uint
	Nombre        string
	CI            string
	Cargo         string
	CodigoUsuario string
	RegionalID    *uint
	SucursalID    *uint
	IsActive      bool
}

func (s *UsuarioService) Actualizar(input ActualizarUsuarioInput) (*models.Usuario, error) {
	usuario, err := s.repo.GetByID(input.ID)
	if err != nil {
		return nil, err
	}

	usuario.Nombre = input.Nombre
	usuario.CI = input.CI
	usuario.Cargo = input.Cargo
	usuario.CodigoUsuario = input.CodigoUsuario
	usuario.RegionalID = input.RegionalID
	usuario.SucursalID = input.SucursalID
	usuario.IsActive = input.IsActive

	if err := s.repo.Update(usuario); err != nil {
		return nil, err
	}
	return usuario, nil
}

// ResetPassword restablece la contraseña del usuario a su CI actual.
func (s *UsuarioService) ResetPassword(id uint) error {
	usuario, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	passwordHash, err := hashPassword(usuario.CI)
	if err != nil {
		return err
	}
	usuario.PasswordHash = passwordHash
	return s.repo.Update(usuario)
}

func (s *UsuarioService) Eliminar(id uint) error {
	return s.repo.SoftDelete(id)
}

func (s *UsuarioService) ObtenerPorID(id uint) (*models.Usuario, error) {
	return s.repo.GetByID(id)
}

func (s *UsuarioService) ListarTodos() ([]models.Usuario, error) {
	return s.repo.GetAll()
}

func (s *UsuarioService) ListarRegionales() ([]models.Regional, error) {
	return s.repo.GetRegionales()
}

func (s *UsuarioService) ListarSucursalesCatalogo() ([]models.SucursalCatalogo, error) {
	return s.repo.GetSucursalesCatalogo()
}

type AccesosInput struct {
	AccesoTotal   bool
	RegionalesIDs []uint
	SucursalesIDs []uint
}

func (s *UsuarioService) ConfigurarAccesos(usuarioID uint, input AccesosInput) error {
	return s.repo.SetAccesos(usuarioID, input.AccesoTotal, input.RegionalesIDs, input.SucursalesIDs)
}

func (s *UsuarioService) ObtenerAccesos(usuarioID uint) (*repositories.AccesosResueltos, error) {
	return s.repo.GetAccesos(usuarioID)
}

// SucursalesPermitidas resuelve qué sucursales del catálogo puede ver un
// usuario (todo el catálogo si tiene acceso total).
func (s *UsuarioService) SucursalesPermitidas(usuarioID uint) ([]models.SucursalCatalogo, error) {
	return s.repo.SucursalesPermitidas(usuarioID)
}

// TieneAccesoSucursal verifica si el usuario puede acceder a la sucursal
// identificada por su código SIN — usado al filtrar consultas de facturas
// por sucursal (ver ConsultasHandler.DataFacturas).
func (s *UsuarioService) TieneAccesoSucursal(usuarioID uint, codigoSucursalSin int) (bool, error) {
	return s.repo.TieneAccesoSucursal(usuarioID, codigoSucursalSin)
}

// TieneAccesoTotal indica si el usuario tiene acceso nacional (bypass de
// sucursales_permitidas_codigos).
func (s *UsuarioService) TieneAccesoTotal(usuarioID uint) (bool, error) {
	return s.repo.TieneAccesoTotal(usuarioID)
}

// ErrCredencialesInvalidas se devuelve cuando el código de usuario no existe,
// la contraseña no coincide, o el usuario está inactivo — mismo mensaje
// genérico en los 3 casos para no filtrar cuáles códigos de usuario existen.
var ErrCredencialesInvalidas = errors.New("usuario o contraseña incorrectos")

// Login verifica codigo_usuario + password contra el hash guardado.
func (s *UsuarioService) Login(codigoUsuario, password string) (*models.Usuario, error) {
	usuario, err := s.repo.GetByCodigoUsuario(codigoUsuario)
	if err != nil {
		return nil, ErrCredencialesInvalidas
	}
	if !usuario.IsActive {
		return nil, ErrCredencialesInvalidas
	}
	if err := bcrypt.CompareHashAndPassword([]byte(usuario.PasswordHash), []byte(password)); err != nil {
		return nil, ErrCredencialesInvalidas
	}
	return usuario, nil
}
