package main

import (
	"fmt"
	"log"
	"managerfact/aplication/services"
	"managerfact/infraestructura/handlers"
	"managerfact/internal/domain/models"
	"managerfact/internal/domain/repositories"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// Config estructura de configuración
type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	DBSSLMode  string
	ServerPort string
}

// LoadConfig carga la configuración desde variables de entorno
func LoadConfig() *Config {
	// Cargar archivo .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
	}

	config := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "invoices_system"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}

	return config
}

// getEnv obtiene una variable de entorno o retorna un valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// InitDatabase inicializa la conexión a PostgreSQL
func InitDatabase(config *Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=America/La_Paz",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort, config.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}

	log.Println("Conexión a PostgreSQL establecida exitosamente")
	return db
}

// AutoMigrate ejecuta las migraciones automáticas
func AutoMigrate(db *gorm.DB) error {
	log.Println("Ejecutando migraciones automáticas...")

	err := db.AutoMigrate(
		&models.DbConnection{},
		&models.Codigo_producto{},
	)

	if err != nil {
		return fmt.Errorf("error en migración automática: %v", err)
	}

	log.Println("Migraciones completadas exitosamente")
	return nil
}

// SeedDatabase agrega datos iniciales si es necesario
func SeedDatabase(db *gorm.DB) error {
	log.Println("Verificando datos iniciales...")

	// Verificar si ya existen conexiones
	var count int64
	if err := db.Model(&models.DbConnection{}).Count(&count).Error; err != nil {
		return fmt.Errorf("error verificando datos existentes: %v", err)
	}

	// Si no hay datos, agregar conexiones de ejemplo (opcional)
	if count == 0 {
		log.Println("No se encontraron conexiones, agregando datos de ejemplo...")

		sampleConnections := []models.DbConnection{
			{
				ServerName:   "Servidor Principal",
				Host:         "localhost",
				Port:         1433,
				DatabaseName: "FacturasDB",
				Username:     "sa",
				Password:     "your_password_here",
				IsActive:     true,
				Description:  "Servidor principal de facturas",
			},
			{
				ServerName:   "Servidor Backup",
				Host:         "backup.example.com",
				Port:         1433,
				DatabaseName: "FacturasDB_Backup",
				Username:     "backup_user",
				Password:     "backup_password_here",
				IsActive:     false,
				Description:  "Servidor de respaldo",
			},
		}

		for _, conn := range sampleConnections {
			if err := db.Create(&conn).Error; err != nil {
				log.Printf("Error creando conexión de ejemplo %s: %v", conn.ServerName, err)
			} else {
				log.Printf("Conexión de ejemplo '%s' creada", conn.ServerName)
			}
		}
	}

	return nil
}

// SetupRoutes configura todas las rutas de la API
func SetupRoutes(
	app *fiber.App,
	dbConnectionHandler *handlers.DbConnectionHandler,
	consultasHandler *handlers.ConsultasHandler,
	codigoProductoHandler *handlers.CodigoProductoHandler,
) {
	// Middleware global
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path} - ${latency}\n",
	}))
	app.Use(recover.New())

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Ruta raíz
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Invoice System API",
			"version": "1.0.0",
			"endpoints": fiber.Map{
				"health":      "/api/v1/health",
				"connections": "/api/v1/connections",
			},
		})
	})

	// Grupo de rutas API
	api := app.Group("/api/v1")

	// Ruta de health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "API funcionando correctamente",
		})
	})

	// Registrar rutas de conexiones de BD
	dbConnectionHandler.RegisterRoutes(api)
	// Registrar rutas de consultas
	consultasHandler.RegisterRoutes(api)
	// Registrar rutas de codigo producto
	codigoProductoHandler.RegisterRoutes(api)
}

func main() {
	log.Println("Iniciando Invoice System API...")

	// Cargar configuración
	config := LoadConfig()

	// Inicializar base de datos
	db := InitDatabase(config)

	// Ejecutar migraciones
	if err := AutoMigrate(db); err != nil {
		log.Fatalf("Error en migraciones: %v", err)
	}

	// Agregar datos iniciales (opcional)
	if err := SeedDatabase(db); err != nil {
		log.Printf("Advertencia en seed de datos: %v", err)
	}

	// Inicializar dependencias (Dependency Injection)
	dbConnectionRepo := repositories.NewDbConnectionRepository(db)
	dbConnectionService := services.NewDbConnectionService(dbConnectionRepo)
	dbConnectionHandler := handlers.NewDbConnectionHandler(dbConnectionService)

	// Iniciar consultas
	consultasRepositori := repositories.NewConsutasRepository(db)
	consultaHandler := services.NewConsultasService(consultasRepositori)
	consultasHandler := handlers.NewConsultasHandler(consultaHandler)

	// codigo producto
	codigoProductoRepo := repositories.NewCodigoProductoRepoRepo(db)
	codigoProductoService := services.NewCodigoProductoService(codigoProductoRepo)
	codigoProductoHandler := handlers.NewCodigoProductoHandler(codigoProductoService)
	// Configurar Fiber
	app := fiber.New(fiber.Config{
		AppName:      "Invoice System API v1.0.0",
		ServerHeader: "Invoice System",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": "Error interno del servidor",
				"error":   err.Error(),
			})
		},
	})

	// Configurar rutas
	SetupRoutes(app, dbConnectionHandler, consultasHandler, codigoProductoHandler)

	// Iniciar servidor
	port := ":" + config.ServerPort
	log.Printf("Servidor ejecutándose en puerto %s", config.ServerPort)
	log.Printf("Endpoints disponibles:")
	log.Printf("  - Health Check: http://localhost%s/api/v1/health", port)
	log.Printf("  - Connections: http://localhost%s/api/v1/connections", port)

	if err := app.Listen(port); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
