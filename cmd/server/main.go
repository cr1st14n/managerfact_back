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
		&models.Regional{},
		&models.SucursalCatalogo{},
		&models.Usuario{},
		&models.UsuarioAccesoRegional{},
		&models.UsuarioAccesoSucursal{},
		&models.SucursalFacturador{},
		&models.FacturaPrevalorada{},
		&models.FacturaAnulacion{},
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

// SeedRegionalesYSucursales siembra el catálogo de regionales y sucursales
// (transcrito de doc/sucursales.md) si todavía no existen datos. Pendiente
// de validar contra SFE_SUCURSAL en producción, tal como advierte el propio
// documento de origen.
func SeedRegionalesYSucursales(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Regional{}).Count(&count).Error; err != nil {
		return fmt.Errorf("error verificando regionales existentes: %v", err)
	}
	if count > 0 {
		return nil
	}

	log.Println("No se encontraron regionales, sembrando catálogo desde doc/sucursales.md...")

	type sucursalSeed struct {
		Codigo int
		Nombre string
	}

	sucursalesPorRegional := map[string][]sucursalSeed{
		"La Paz": {
			{4, "El Alto"}, {5, "Oruro"}, {25, "Cobija"}, {28, "Uyuni"},
			{35, "Rurrenabaque"}, {36, "Reyes"}, {24, "San Borja"},
			{6, "Copacabana"}, {38, "Apolo"},
		},
		"Cochabamba": {
			{3, "Cochabamba"}, {31, "Chimoré"}, {33, "Tarija"}, {26, "Potosí"},
			{37, "Alcantarí"}, {30, "Yacuiba"}, {8, "Monteagudo"},
			{27, "Villamontes"}, {7, "Bermejo"},
		},
		"Santa Cruz": {
			{29, "Viru Viru"}, {2, "Trompillo"}, {17, "San Javier"}, {11, "Concepción"},
			{14, "San Ignacio de Velasco"}, {19, "Camiri"}, {10, "Roboré"},
			{13, "Puerto Suárez"}, {12, "Vallegrande"}, {18, "Ascensión de Guarayos"},
			{16, "San José Chiquitos"}, {9, "San Matías"},
		},
		// San Ramón (Beni) no tiene código SIN visible en la tarjeta fuente,
		// así que queda fuera del catálogo hasta confirmarlo.
		"Beni": {
			{1, "Trinidad"}, {22, "Santa Ana de Yacuma"}, {23, "San Ignacio de Moxos"},
			{21, "Magdalena"}, {34, "Riberalta"}, {32, "Santa Rosa"}, {20, "Guarayamerín"},
		},
	}

	ordenRegionales := []string{"La Paz", "Cochabamba", "Santa Cruz", "Beni"}

	for _, nombreRegional := range ordenRegionales {
		regional := models.Regional{Nombre: nombreRegional}
		if err := db.Create(&regional).Error; err != nil {
			return fmt.Errorf("error creando regional %s: %v", nombreRegional, err)
		}
		for _, s := range sucursalesPorRegional[nombreRegional] {
			sucursal := models.SucursalCatalogo{
				CodigoSucursalSin: s.Codigo,
				Nombre:            s.Nombre,
				RegionalID:        regional.ID,
			}
			if err := db.Create(&sucursal).Error; err != nil {
				return fmt.Errorf("error creando sucursal %s: %v", s.Nombre, err)
			}
		}
	}

	log.Println("Catálogo de regionales y sucursales sembrado exitosamente")
	return nil
}

// SetupRoutes configura todas las rutas de la API
func SetupRoutes(
	app *fiber.App,
	dbConnectionHandler *handlers.DbConnectionHandler,
	consultasHandler *handlers.ConsultasHandler,
	codigoProductoHandler *handlers.CodigoProductoHandler,
	usuarioHandler *handlers.UsuarioHandler,
	sucursalFacturadorHandler *handlers.SucursalFacturadorHandler,
	facturaPrevaloradaHandler *handlers.FacturaPrevaloradaHandler,
	facturaAnulacionHandler *handlers.FacturaAnulacionHandler,
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
	// Registrar rutas de usuarios/regionales/catálogo de sucursales
	usuarioHandler.RegisterRoutes(api)
	// Registrar rutas de sucursales facturador (FacturaClic)
	sucursalFacturadorHandler.RegisterRoutes(api)
	// Registrar rutas de facturas prevaloradas (boletos)
	facturaPrevaloradaHandler.RegisterRoutes(api)
	// Registrar rutas de facturas de anulación
	facturaAnulacionHandler.RegisterRoutes(api)
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
	if err := SeedRegionalesYSucursales(db); err != nil {
		log.Printf("Advertencia en seed de regionales/sucursales: %v", err)
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

	// usuarios
	usuarioRepo := repositories.NewUsuarioRepository(db)
	usuarioService := services.NewUsuarioService(usuarioRepo)
	usuarioHandler := handlers.NewUsuarioHandler(usuarioService)

	// sucursales facturador (FacturaClic)
	sucursalFacturadorRepo := repositories.NewSucursalFacturadorRepository(db)
	sucursalFacturadorService := services.NewSucursalFacturadorService(sucursalFacturadorRepo)
	sucursalFacturadorHandler := handlers.NewSucursalFacturadorHandler(sucursalFacturadorService)

	// facturas prevaloradas (boletos)
	facturaPrevaloradaRepo := repositories.NewFacturaPrevaloradaRepository(db)
	facturaPrevaloradaService := services.NewFacturaPrevaloradaService(facturaPrevaloradaRepo, sucursalFacturadorRepo)
	facturaPrevaloradaHandler := handlers.NewFacturaPrevaloradaHandler(facturaPrevaloradaService)

	// facturas de anulación
	facturaAnulacionRepo := repositories.NewFacturaAnulacionRepository(db)
	facturaAnulacionService := services.NewFacturaAnulacionService(facturaAnulacionRepo, sucursalFacturadorRepo)
	facturaAnulacionHandler := handlers.NewFacturaAnulacionHandler(facturaAnulacionService)
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
	SetupRoutes(app, dbConnectionHandler, consultasHandler, codigoProductoHandler, usuarioHandler, sucursalFacturadorHandler, facturaPrevaloradaHandler, facturaAnulacionHandler)

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
