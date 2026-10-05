package models

import (
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

)

const (
	AmbienteProduccion = "produccion"
	AmbienteBaja       = "baja"
	AmbienteTest       = "test"
)

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

type DbConnection struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	ServerName   string `json:"server_name" gorm:"type:varchar(100);not null;uniqueIndex" validate:"required,min=3,max=100"`
	Host         string `json:"host" gorm:"type:varchar(255);not null" validate:"required,hostname_rfc1123"`
	Port         int    `json:"port" gorm:"not null;default:1433" validate:"required,min=1,max=65535"`
	DatabaseName string `json:"database_name" gorm:"type:varchar(100);not null" validate:"required,min=1,max=100"`
	Username     string `json:"username" gorm:"type:varchar(100);not null" validate:"required,min=1,max=100"`

	Password    string `json:"-" gorm:"type:varchar(255);not null" validate:"required,min=1"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	Description string `json:"description" gorm:"type:text"`
	Type        string `json:"type" gorm:"column:type;type:varchar(50);not null;default:'general'" validate:"required,min=1,max=50"`

	Ambiente  string         `json:"ambiente" gorm:"type:varchar(20);not null;default:'produccion'" validate:"required,oneof=produccion baja test"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (DbConnection) TableName() string {
	return "db_connections"
}

func (dc DbConnection) MarshalJSON() ([]byte, error) {
	type alias DbConnection
	return json.Marshal(struct {
		alias
		PasswordConfigurado bool `json:"password_configurado"`
	}{alias(dc), dc.Password != ""})
}

func (dc *DbConnection) BeforeCreate(tx *gorm.DB) error {

	return nil
}

func (dc *DbConnection) dsnURL() *url.URL {
	return &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(dc.Username, dc.Password),
		Host:     net.JoinHostPort(dc.Host, strconv.Itoa(dc.Port)),
		RawQuery: url.Values{"database": {dc.DatabaseName}}.Encode(),
	}
}

func (dc *DbConnection) DSN() string {
	return dc.dsnURL().String()
}

func (dc *DbConnection) ConnectionString() string {
	u := dc.dsnURL()
	q := u.Query()
	q.Set("encrypt", "false")
	q.Set("connection timeout", "30")
	u.RawQuery = q.Encode()
	return u.String()
}

func (dc *DbConnection) IsValid() bool {
	return dc.ServerName != "" &&
		dc.Host != "" &&
		dc.Port > 0 &&
		dc.DatabaseName != "" &&
		dc.Username != "" &&
		dc.Password != ""
}
