package services

import (
	"errors"
	"fmt"
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"

	"gorm.io/gorm"
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

func (s *CodigoProductoService) ObtenerPorID(id uint) (*models.Codigo_producto, error) {
	data, err := s.CodigoProductoRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("código de producto con ID %d no encontrado", id)
		}
		return nil, err
	}
	return data, nil
}

func (s *CodigoProductoService) Crear(codigo, descripcion string) (*models.Codigo_producto, error) {
	existe, err := s.CodigoProductoRepo.ExisteCodigo(codigo, 0)
	if err != nil {
		return nil, err
	}
	if existe {
		return nil, fmt.Errorf("ya existe un código de producto '%s'", codigo)
	}

	data := &models.Codigo_producto{Codigo: codigo, Descripcion: descripcion}
	if err := s.CodigoProductoRepo.Create(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *CodigoProductoService) Actualizar(id uint, codigo, descripcion string) (*models.Codigo_producto, error) {
	data, err := s.ObtenerPorID(id)
	if err != nil {
		return nil, err
	}

	existe, err := s.CodigoProductoRepo.ExisteCodigo(codigo, id)
	if err != nil {
		return nil, err
	}
	if existe {
		return nil, fmt.Errorf("ya existe otro código de producto '%s'", codigo)
	}

	data.Codigo = codigo
	data.Descripcion = descripcion
	if err := s.CodigoProductoRepo.Update(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *CodigoProductoService) Eliminar(id uint) error {
	if err := s.CodigoProductoRepo.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("código de producto con ID %d no encontrado", id)
		}
		return err
	}
	return nil
}
