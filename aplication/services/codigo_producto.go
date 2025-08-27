package services

import (
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"
)

type CodigoProductoService struct {
	CodigoProductoRepo *repositories.CodigoProductoRepo
}

func NewCodigoProductoService(r *repositories.CodigoProductoRepo) *CodigoProductoService {
	return &CodigoProductoService{
		CodigoProductoRepo: r,
	}
}
func (s *CodigoProductoService) Get() (*[]models.Codigo_producto, error) {
	data, err := s.CodigoProductoRepo.GetAll()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *CodigoProductoService) GetByCodigo(codigo string) (*models.Codigo_producto, error) {
	data, err := s.CodigoProductoRepo.GetByCodigo(codigo)
	if err != nil {
		return nil, err
	}
	return data, nil
}
