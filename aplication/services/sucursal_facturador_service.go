package services

import (
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"
	"managerfact/pkg/utils"
)

type SucursalFacturadorService struct {
	repo *repositories.SucursalFacturadorRepository
}

func NewSucursalFacturadorService(r *repositories.SucursalFacturadorRepository) *SucursalFacturadorService {
	return &SucursalFacturadorService{repo: r}
}

type CrearSucursalFacturadorInput struct {
	Nombre            string
	CodigoSucursalSin int
	PuntoVentaEmisor  string
	UrlLinkFacturador string
	TokenAcceso       string
	CodigoMonedaBob   string
	CodigoCI          string
	CodigoNit         string
}

func (s *SucursalFacturadorService) Crear(input CrearSucursalFacturadorInput) (*models.SucursalFacturador, error) {
	tokenCifrado, err := utils.Encrypt(input.TokenAcceso)
	if err != nil {
		return nil, err
	}

	sucursal := &models.SucursalFacturador{
		Nombre:            input.Nombre,
		CodigoSucursalSin: input.CodigoSucursalSin,
		PuntoVentaEmisor:  input.PuntoVentaEmisor,
		UrlLinkFacturador: input.UrlLinkFacturador,
		TokenAcceso:       tokenCifrado,
		CodigoMonedaBob:   input.CodigoMonedaBob,
		CodigoCI:          input.CodigoCI,
		CodigoNit:         input.CodigoNit,
		Activo:            true,
	}
	if err := s.repo.Create(sucursal); err != nil {
		return nil, err
	}
	return sucursal, nil
}

type ActualizarSucursalFacturadorInput struct {
	ID                uint
	Nombre            string
	CodigoSucursalSin int
	PuntoVentaEmisor  string
	UrlLinkFacturador string
	// TokenAcceso solo se re-cifra y actualiza si viene con valor; el
	// formulario de edición nunca recibe el token actual de vuelta (no se
	// expone por la API), así que un campo vacío significa "no cambiar".
	TokenAcceso     string
	CodigoMonedaBob string
	CodigoCI        string
	CodigoNit       string
	Activo          bool
}

func (s *SucursalFacturadorService) Actualizar(input ActualizarSucursalFacturadorInput) (*models.SucursalFacturador, error) {
	sucursal, err := s.repo.GetByID(input.ID)
	if err != nil {
		return nil, err
	}

	sucursal.Nombre = input.Nombre
	sucursal.CodigoSucursalSin = input.CodigoSucursalSin
	sucursal.PuntoVentaEmisor = input.PuntoVentaEmisor
	sucursal.UrlLinkFacturador = input.UrlLinkFacturador
	sucursal.CodigoMonedaBob = input.CodigoMonedaBob
	sucursal.CodigoCI = input.CodigoCI
	sucursal.CodigoNit = input.CodigoNit
	sucursal.Activo = input.Activo

	if input.TokenAcceso != "" {
		tokenCifrado, err := utils.Encrypt(input.TokenAcceso)
		if err != nil {
			return nil, err
		}
		sucursal.TokenAcceso = tokenCifrado
	}

	if err := s.repo.Update(sucursal); err != nil {
		return nil, err
	}
	return sucursal, nil
}

func (s *SucursalFacturadorService) Eliminar(id uint) error {
	return s.repo.SoftDelete(id)
}

func (s *SucursalFacturadorService) ObtenerPorID(id uint) (*models.SucursalFacturador, error) {
	return s.repo.GetByID(id)
}

func (s *SucursalFacturadorService) ListarTodos() ([]models.SucursalFacturador, error) {
	return s.repo.GetAll()
}

// TokenDescifrado descifra el token en memoria, para uso exclusivo del
// futuro FacturadorClient al armar la request HTTP saliente.
func (s *SucursalFacturadorService) TokenDescifrado(id uint) (string, error) {
	sucursal, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	return utils.Decrypt(sucursal.TokenAcceso)
}
