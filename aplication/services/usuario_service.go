package services

import (
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

func (s *UsuarioService) ObtenerAccesos(usuarioID uint) ([]models.UsuarioAccesoRegional, []models.UsuarioAccesoSucursal, error) {
	return s.repo.GetAccesos(usuarioID)
}

// SucursalesPermitidas resuelve qué sucursales del catálogo puede ver un
// usuario. Todavía no está conectado a ningún flujo de login/sesión: es el
// endpoint que se usará cuando eso exista.
func (s *UsuarioService) SucursalesPermitidas(usuarioID uint) ([]models.SucursalCatalogo, error) {
	return s.repo.SucursalesPermitidas(usuarioID)
}
