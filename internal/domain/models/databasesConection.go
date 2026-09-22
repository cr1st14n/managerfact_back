package models

// package models

import (
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	// "gorm.io/gorm"
)

// Ambientes de una conexión. En la UI se muestran con color: producción
// (verde), baja (amarillo) y test (rojo).
const (
	AmbienteProduccion = "produccion"
	AmbienteBaja       = "baja"
	AmbienteTest       = "test"
)

// NormalizarAmbiente limpia el valor recibido y verifica que sea uno de los
// ambientes válidos. Vacío se toma como producción; quien necesite conservar
// el valor guardado cuando no viene (editar) debe revisarlo antes.
func NormalizarAmbiente(ambiente string) (string, bool) {
	ambiente = strings.ToLower(strings.TrimSpace(ambiente))
	switch ambiente {
	case "":
		return AmbienteProduccion, true
	case AmbienteProduccion, AmbienteBaja, AmbienteTest:
		return ambiente, true
	}
	return "", false
}

// DbConnection representa la configuración de conexión a bases de datos SQL Server
type DbConnection struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	ServerName   string `json:"server_name" gorm:"type:varchar(100);not null;uniqueIndex" validate:"required,min=3,max=100"`
	Host         string `json:"host" gorm:"type:varchar(255);not null" validate:"required,hostname_rfc1123"`
	Port         int    `json:"port" gorm:"not null;default:1433" validate:"required,min=1,max=65535"`
	DatabaseName string `json:"database_name" gorm:"type:varchar(100);not null" validate:"required,min=1,max=100"`
	Username     string `json:"username" gorm:"type:varchar(100);not null" validate:"required,min=1,max=100"`
	// Password nunca se serializa: la API solo expone "password_configurado"
	// (ver MarshalJSON). Se recibe únicamente al crear la conexión.
	Password    string `json:"-" gorm:"type:varchar(255);not null" validate:"required,min=1"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	Description string `json:"description" gorm:"type:text"`
	Type        string `json:"type" gorm:"column:type;type:varchar(50);not null;default:'general'" validate:"required,min=1,max=50"`
	// Ambiente: las conexiones existentes quedan como producción al agregarse
	// la columna (default); el admin las reclasifica desde /conexiones.
	Ambiente  string         `json:"ambiente" gorm:"type:varchar(20);not null;default:'produccion'" validate:"required,oneof=produccion baja test"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName especifica el nombre de la tabla
func (DbConnection) TableName() string {
	return "db_connections"
}

// MarshalJSON agrega "password_configurado" en lugar de la contraseña. Va en
// el modelo (y no en un wrapper del handler) porque el listado que consumen
// los selectores de base devuelve []DbConnection directo desde el servicio.
func (dc DbConnection) MarshalJSON() ([]byte, error) {
	type alias DbConnection
	return json.Marshal(struct {
		alias
		PasswordConfigurado bool `json:"password_configurado"`
	}{alias(dc), dc.Password != ""})
}

// BeforeCreate se ejecuta antes de crear un registro
func (dc *DbConnection) BeforeCreate(tx *gorm.DB) error {
	// Aquí se podría encriptar la contraseña si es necesario
	return nil
}

// dsnURL arma la URL de conexión de SQL Server. Usuario y contraseña van
// escapados con url.UserPassword: armarla con Sprintf hacía que una
// contraseña con "@", ":", "/" o "#" rompiera la conexión.
func (dc *DbConnection) dsnURL() *url.URL {
	return &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(dc.Username, dc.Password),
		Host:     net.JoinHostPort(dc.Host, strconv.Itoa(dc.Port)),
		RawQuery: url.Values{"database": {dc.DatabaseName}}.Encode(),
	}
}

// DSN es la cadena que usan las consultas contra el SQL Server de la
// conexión (sin parámetros extra: igual que antes, con encriptación por
// defecto del driver).
func (dc *DbConnection) DSN() string {
	return dc.dsnURL().String()
}

// ConnectionString es la cadena de la prueba de conexión: el mismo DSN, con
// encrypt=false y timeout de 30 s.
func (dc *DbConnection) ConnectionString() string {
	u := dc.dsnURL()
	q := u.Query()
	q.Set("encrypt", "false")
	q.Set("connection timeout", "30")
	u.RawQuery = q.Encode()
	return u.String()
}

// IsValid verifica si la configuración tiene los campos requeridos
func (dc *DbConnection) IsValid() bool {
	return dc.ServerName != "" &&
		dc.Host != "" &&
		dc.Port > 0 &&
		dc.DatabaseName != "" &&
		dc.Username != "" &&
		dc.Password != ""
}
