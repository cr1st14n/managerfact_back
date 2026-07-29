package repositories

import (
	"errors"
	"fmt"
	"managerfact/internal/domain/models"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type UsuarioRepository struct {
	db *gorm.DB
}

func NewUsuarioRepository(db *gorm.DB) *UsuarioRepository {
	return &UsuarioRepository{db: db}
}

func (r *UsuarioRepository) Create(usuario *models.Usuario) error {
	if err := r.db.Create(usuario).Error; err != nil {
		return fmt.Errorf("error creando usuario: %w", err)
	}
	return nil
}

func (r *UsuarioRepository) GetByID(id uint) (*models.Usuario, error) {
	var usuario models.Usuario
	err := r.db.Preload("Regional").Preload("Sucursal").First(&usuario, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("usuario con ID %d no encontrado", id)
		}
		return nil, fmt.Errorf("error obteniendo usuario: %w", err)
	}
	return &usuario, nil
}

// GetByCodigoUsuario busca un usuario por su código de acceso (login);
// usado por UsuarioService.Login.
func (r *UsuarioRepository) GetByCodigoUsuario(codigoUsuario string) (*models.Usuario, error) {
	var usuario models.Usuario
	err := r.db.Where("codigo_usuario = ?", codigoUsuario).First(&usuario).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("usuario %q no encontrado", codigoUsuario)
		}
		return nil, fmt.Errorf("error obteniendo usuario: %w", err)
	}
	return &usuario, nil
}

func (r *UsuarioRepository) GetAll() ([]models.Usuario, error) {
	var usuarios []models.Usuario
	err := r.db.Preload("Regional").Preload("Sucursal").Order("nombre ASC").Find(&usuarios).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo usuarios: %w", err)
	}
	return usuarios, nil
}

func (r *UsuarioRepository) Update(usuario *models.Usuario) error {
	if err := r.db.Save(usuario).Error; err != nil {
		return fmt.Errorf("error actualizando usuario: %w", err)
	}
	return nil
}

func (r *UsuarioRepository) SoftDelete(id uint) error {
	result := r.db.Delete(&models.Usuario{}, id)
	if result.Error != nil {
		return fmt.Errorf("error eliminando usuario: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("usuario con ID %d no encontrado", id)
	}
	return nil
}

// GetRegionales lista las regionales disponibles (La Paz, Cochabamba, Santa
// Cruz, Beni).
func (r *UsuarioRepository) GetRegionales() ([]models.Regional, error) {
	var regionales []models.Regional
	err := r.db.Order("nombre ASC").Find(&regionales).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo regionales: %w", err)
	}
	return regionales, nil
}

// GetSucursalesCatalogo lista el catálogo maestro de sucursales.
func (r *UsuarioRepository) GetSucursalesCatalogo() ([]models.SucursalCatalogo, error) {
	var sucursales []models.SucursalCatalogo
	err := r.db.Preload("Regional").Order("codigo_sucursal_sin ASC").Find(&sucursales).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo catálogo de sucursales: %w", err)
	}
	return sucursales, nil
}

// codigosATexto ordena, deduplica y concatena códigos SIN como "1,6,7".
func codigosATexto(codigos []int) string {
	if len(codigos) == 0 {
		return ""
	}
	set := make(map[int]struct{}, len(codigos))
	for _, c := range codigos {
		set[c] = struct{}{}
	}
	unicos := make([]int, 0, len(set))
	for c := range set {
		unicos = append(unicos, c)
	}
	sort.Ints(unicos)
	partes := make([]string, len(unicos))
	for i, c := range unicos {
		partes[i] = strconv.Itoa(c)
	}
	return strings.Join(partes, ",")
}

// textoACodigos parsea "1,6,7" a un set de códigos SIN. Texto vacío = sin
// accesos.
func textoACodigos(texto string) map[int]struct{} {
	set := map[int]struct{}{}
	texto = strings.TrimSpace(texto)
	if texto == "" {
		return set
	}
	for _, parte := range strings.Split(texto, ",") {
		parte = strings.TrimSpace(parte)
		if parte == "" {
			continue
		}
		if codigo, err := strconv.Atoi(parte); err == nil {
			set[codigo] = struct{}{}
		}
	}
	return set
}

// SetAccesos reemplaza por completo la configuración de accesos de un
// usuario: la bandera de acceso total y la colección de códigos SIN de
// sucursal permitidos. regionalesIDs/sucursalesIDs son selección por
// conveniencia desde el front (agrupada visualmente); acá se resuelven a
// los códigos SIN de sus sucursales y se guardan ya aplanados en
// sucursales_permitidas_codigos.
func (r *UsuarioRepository) SetAccesos(usuarioID uint, accesoTotal bool, regionalesIDs, sucursalesIDs []uint) error {
	var codigos []int

	if len(regionalesIDs) > 0 {
		var codigosRegionales []int
		if err := r.db.Model(&models.SucursalCatalogo{}).
			Where("regional_id IN ?", regionalesIDs).
			Pluck("codigo_sucursal_sin", &codigosRegionales).Error; err != nil {
			return fmt.Errorf("error resolviendo sucursales de las regionales: %w", err)
		}
		codigos = append(codigos, codigosRegionales...)
	}

	if len(sucursalesIDs) > 0 {
		var codigosDirectos []int
		if err := r.db.Model(&models.SucursalCatalogo{}).
			Where("id IN ?", sucursalesIDs).
			Pluck("codigo_sucursal_sin", &codigosDirectos).Error; err != nil {
			return fmt.Errorf("error resolviendo códigos de las sucursales: %w", err)
		}
		codigos = append(codigos, codigosDirectos...)
	}

	err := r.db.Model(&models.Usuario{}).Where("id = ?", usuarioID).Updates(map[string]interface{}{
		"acceso_total":                  accesoTotal,
		"sucursales_permitidas_codigos": codigosATexto(codigos),
	}).Error
	if err != nil {
		return fmt.Errorf("error guardando accesos: %w", err)
	}
	return nil
}

// AccesosResueltos es la selección por regional/sucursal derivada de la
// colección de códigos SIN guardada — reconstruye lo que el front necesita
// para precargar el modal de accesos (qué regionales aparecen "completas" y
// qué sucursales sueltas quedan marcadas).
type AccesosResueltos struct {
	RegionalesIDs []uint
	SucursalesIDs []uint
}

// GetAccesos deriva, a partir de sucursales_permitidas_codigos, qué
// regionales están completamente cubiertas y qué sucursales sueltas quedan
// marcadas individualmente.
func (r *UsuarioRepository) GetAccesos(usuarioID uint) (*AccesosResueltos, error) {
	usuario, err := r.GetByID(usuarioID)
	if err != nil {
		return nil, err
	}
	permitidos := textoACodigos(usuario.SucursalesPermitidasCodigos)

	sucursales, err := r.GetSucursalesCatalogo()
	if err != nil {
		return nil, err
	}

	porRegional := map[uint][]models.SucursalCatalogo{}
	for _, s := range sucursales {
		porRegional[s.RegionalID] = append(porRegional[s.RegionalID], s)
	}

	resultado := &AccesosResueltos{RegionalesIDs: []uint{}, SucursalesIDs: []uint{}}
	for regionalID, lista := range porRegional {
		completa := len(lista) > 0
		for _, s := range lista {
			if _, ok := permitidos[s.CodigoSucursalSin]; !ok {
				completa = false
				break
			}
		}
		if completa {
			resultado.RegionalesIDs = append(resultado.RegionalesIDs, regionalID)
			continue
		}
		for _, s := range lista {
			if _, ok := permitidos[s.CodigoSucursalSin]; ok {
				resultado.SucursalesIDs = append(resultado.SucursalesIDs, s.ID)
			}
		}
	}
	sort.Slice(resultado.RegionalesIDs, func(i, j int) bool { return resultado.RegionalesIDs[i] < resultado.RegionalesIDs[j] })
	sort.Slice(resultado.SucursalesIDs, func(i, j int) bool { return resultado.SucursalesIDs[i] < resultado.SucursalesIDs[j] })
	return resultado, nil
}

// SucursalesPermitidas resuelve el catálogo efectivo de sucursales a las que
// el usuario tiene acceso: todo el catálogo si tiene AccesoTotal, o las que
// coincidan con sucursales_permitidas_codigos.
func (r *UsuarioRepository) SucursalesPermitidas(usuarioID uint) ([]models.SucursalCatalogo, error) {
	usuario, err := r.GetByID(usuarioID)
	if err != nil {
		return nil, err
	}

	if usuario.AccesoTotal {
		return r.GetSucursalesCatalogo()
	}

	permitidos := textoACodigos(usuario.SucursalesPermitidasCodigos)
	if len(permitidos) == 0 {
		return []models.SucursalCatalogo{}, nil
	}

	codigos := make([]int, 0, len(permitidos))
	for c := range permitidos {
		codigos = append(codigos, c)
	}

	var sucursales []models.SucursalCatalogo
	if err := r.db.Preload("Regional").
		Where("codigo_sucursal_sin IN ?", codigos).
		Order("codigo_sucursal_sin ASC").
		Find(&sucursales).Error; err != nil {
		return nil, fmt.Errorf("error resolviendo sucursales permitidas: %w", err)
	}
	return sucursales, nil
}

// TieneAccesoTotal indica si el usuario tiene acceso nacional (bypass de
// sucursales_permitidas_codigos).
func (r *UsuarioRepository) TieneAccesoTotal(usuarioID uint) (bool, error) {
	usuario, err := r.GetByID(usuarioID)
	if err != nil {
		return false, err
	}
	return usuario.AccesoTotal, nil
}

// TieneAccesoSucursal verifica si el usuario puede acceder a la sucursal
// identificada por su código SIN — el chequeo real usado al hacer consultas
// de facturas de una sucursal puntual.
func (r *UsuarioRepository) TieneAccesoSucursal(usuarioID uint, codigoSucursalSin int) (bool, error) {
	usuario, err := r.GetByID(usuarioID)
	if err != nil {
		return false, err
	}
	if usuario.AccesoTotal {
		return true, nil
	}
	permitidos := textoACodigos(usuario.SucursalesPermitidasCodigos)
	_, ok := permitidos[codigoSucursalSin]
	return ok, nil
}
