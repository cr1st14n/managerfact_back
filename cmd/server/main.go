package main

import (
	"fmt"
	"log"
	"managerfact/aplication/services"
	"managerfact/infraestructura/handlers"
	"managerfact/infraestructura/middleware"
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

type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	DBSSLMode  string
	ServerPort string
}

func LoadConfig() *Config {

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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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

// La migración usa AT TIME ZONE 'UTC': el cast implícito usa UTC-4 y desplaza un día las fechas importadas.
func MigrarFechasPrevaloradaADate(db *gorm.DB) error {
	var tipoActual string
	err := db.Raw(`
		SELECT data_type FROM information_schema.columns
		WHERE table_name = 'facturas_prevaloradas' AND column_name = 'fecha_emision'
	`).Scan(&tipoActual).Error
	if err != nil {
		return fmt.Errorf("error verificando tipo de fecha_emision: %v", err)
	}
	if tipoActual != "timestamp with time zone" {
		return nil
	}

	log.Println("Migrando fecha_emision/fecha_compra_boleto de facturas_prevaloradas a date (evitando el corrimiento de un día)...")
	err = db.Exec(`
		ALTER TABLE facturas_prevaloradas
			ALTER COLUMN fecha_emision TYPE date USING (fecha_emision AT TIME ZONE 'UTC')::date,
			ALTER COLUMN fecha_compra_boleto TYPE date USING (fecha_compra_boleto AT TIME ZONE 'UTC')::date
	`).Error
	if err != nil {
		return fmt.Errorf("error migrando fechas de facturas_prevaloradas a date: %v", err)
	}
	log.Println("Migración de fechas a date completada")
	return nil
}

func AutoMigrate(db *gorm.DB) error {
	if err := MigrarFechasPrevaloradaADate(db); err != nil {
		return err
	}

	log.Println("Ejecutando migraciones automáticas...")

	err := db.AutoMigrate(
		&models.DbConnection{},
		&models.Codigo_producto{},
		&models.ConexionSucursal{},
		&models.Regional{},
		&models.SucursalCatalogo{},
		&models.Usuario{},
		&models.SucursalFacturador{},
		&models.FacturaPrevalorada{},
		&models.FacturaAnulacion{},
		&models.LogEnvio{},
	)

	if err != nil {
		return fmt.Errorf("error en migración automática: %v", err)
	}

	log.Println("Migraciones completadas exitosamente")
	return nil
}

func SeedDatabase(db *gorm.DB) error {
	log.Println("Verificando datos iniciales...")

	var count int64
	if err := db.Model(&models.DbConnection{}).Count(&count).Error; err != nil {
		return fmt.Errorf("error verificando datos existentes: %v", err)
	}

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

func SeedRegionalesYSucursales(db *gorm.DB) error {
	type sucursalSeed struct {
		Codigo int
		Nombre string
	}

	sucursalesPorRegional := map[string][]sucursalSeed{

		"La Paz": {
			{4, "El Alto"}, {5, "Oruro"}, {25, "Cobija"}, {28, "Uyuni"},
			{35, "Rurrenabaque"}, {36, "Reyes"}, {24, "San Borja"},
			{6, "Copacabana"}, {38, "Apolo"}, {0, "Oficina Central"},
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

		"Beni": {
			{1, "Trinidad"}, {22, "Santa Ana de Yacuma"}, {23, "San Ignacio de Moxos"},
			{21, "Magdalena"}, {34, "Riberalta"}, {32, "Santa Rosa"}, {20, "Guarayamerín"},
		},
	}

	ordenRegionales := []string{"La Paz", "Cochabamba", "Santa Cruz", "Beni"}

	for _, nombreRegional := range ordenRegionales {
		regional := models.Regional{Nombre: nombreRegional}
		if err := db.Where(models.Regional{Nombre: nombreRegional}).FirstOrCreate(&regional).Error; err != nil {
			return fmt.Errorf("error creando regional %s: %v", nombreRegional, err)
		}
		for _, s := range sucursalesPorRegional[nombreRegional] {
			sucursal := models.SucursalCatalogo{
				CodigoSucursalSin: s.Codigo,
				Nombre:            s.Nombre,
				RegionalID:        regional.ID,
			}

			if err := db.Where("codigo_sucursal_sin = ?", s.Codigo).FirstOrCreate(&sucursal).Error; err != nil {
				return fmt.Errorf("error creando sucursal %s: %v", s.Nombre, err)
			}
		}
	}

	log.Println("Catálogo de regionales y sucursales verificado/completado")
	return nil
}

func SetupRoutes(
	app *fiber.App,
	authHandler *handlers.AuthHandler,
	usuarioService *services.UsuarioService,
	dbConnectionHandler *handlers.DbConnectionHandler,
	consultasHandler *handlers.ConsultasHandler,
	codigoProductoHandler *handlers.CodigoProductoHandler,
	conexionSucursalHandler *handlers.ConexionSucursalHandler,
	usuarioHandler *handlers.UsuarioHandler,
	sucursalFacturadorHandler *handlers.SucursalFacturadorHandler,
	facturaPrevaloradaHandler *handlers.FacturaPrevaloradaHandler,
	facturaAnulacionHandler *handlers.FacturaAnulacionHandler,
	logEnvioHandler *handlers.LogEnvioHandler,
	resumenContableHandler *handlers.ResumenContableHandler,
) {

	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path} - ${latency}\n",
	}))
	app.Use(recover.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

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

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "API funcionando correctamente",
		})
	})

	authHandler.RegisterRoutes(api)

	protegido := api.Group("/", middleware.RequireAuth())

	requireAdmin := middleware.RequireAdmin(usuarioService)

	requireNoConsultas := middleware.RequireNoConsultas(usuarioService)

	dbConnectionHandler.RegisterRoutes(protegido, requireAdmin)

	conexionSucursalHandler.RegisterRoutes(protegido, requireAdmin)

	consultasHandler.RegisterRoutes(protegido)
	resumenContableHandler.RegisterRoutes(protegido)

	codigoProductoHandler.RegisterRoutes(protegido, requireAdmin)

	usuarioHandler.RegisterRoutes(protegido, requireAdmin)

	sucursalFacturadorHandler.RegisterRoutes(protegido, requireAdmin, requireNoConsultas)

	facturaPrevaloradaHandler.RegisterRoutes(protegido, requireNoConsultas)

	facturaAnulacionHandler.RegisterRoutes(protegido, requireNoConsultas)

	logEnvioHandler.RegisterRoutes(protegido, requireNoConsultas)
}

func main() {
	log.Println("Iniciando Invoice System API...")

	config := LoadConfig()

	db := InitDatabase(config)

	if err := AutoMigrate(db); err != nil {
		log.Fatalf("Error en migraciones: %v", err)
	}

	if err := SeedDatabase(db); err != nil {
		log.Printf("Advertencia en seed de datos: %v", err)
	}
	if err := SeedRegionalesYSucursales(db); err != nil {
		log.Printf("Advertencia en seed de regionales/sucursales: %v", err)
	}

	dbConnectionRepo := repositories.NewDbConnectionRepository(db)
	dbConnectionService := services.NewDbConnectionService(dbConnectionRepo)
	dbConnectionHandler := handlers.NewDbConnectionHandler(dbConnectionService)

	usuarioRepo := repositories.NewUsuarioRepository(db)
	usuarioService := services.NewUsuarioService(usuarioRepo)
	usuarioHandler := handlers.NewUsuarioHandler(usuarioService)

	authHandler := handlers.NewAuthHandler(usuarioService)

	consultasRepositori := repositories.NewConsutasRepository(db)
	conexionSucursalRepo := repositories.NewConexionSucursalRepo(db)
	conexionSucursalService := services.NewConexionSucursalService(dbConnectionRepo, conexionSucursalRepo)
	conexionSucursalHandler := handlers.NewConexionSucursalHandler(conexionSucursalService)
	consultaHandler := services.NewConsultasService(consultasRepositori, usuarioRepo, conexionSucursalRepo)
	consultasHandler := handlers.NewConsultasHandler(consultaHandler, usuarioService)

	resumenContableService := services.NewResumenContableService(consultasRepositori)
	resumenContableHandler := handlers.NewResumenContableHandler(resumenContableService, usuarioService)

	codigoProductoRepo := repositories.NewCodigoProductoRepoRepo(db)
	codigoProductoService := services.NewCodigoProductoService(codigoProductoRepo)
	codigoProductoHandler := handlers.NewCodigoProductoHandler(codigoProductoService)

	sucursalFacturadorRepo := repositories.NewSucursalFacturadorRepository(db)
	sucursalFacturadorService := services.NewSucursalFacturadorService(sucursalFacturadorRepo)
	sucursalFacturadorHandler := handlers.NewSucursalFacturadorHandler(sucursalFacturadorService)

	logEnvioRepo := repositories.NewLogEnvioRepository(db)
	logEnvioHandler := handlers.NewLogEnvioHandler(logEnvioRepo)

	facturaPrevaloradaRepo := repositories.NewFacturaPrevaloradaRepository(db)
	facturaPrevaloradaService := services.NewFacturaPrevaloradaService(facturaPrevaloradaRepo, sucursalFacturadorRepo, logEnvioRepo, usuarioService)
	facturaPrevaloradaHandler := handlers.NewFacturaPrevaloradaHandler(facturaPrevaloradaService)

	facturaAnulacionRepo := repositories.NewFacturaAnulacionRepository(db)
	facturaAnulacionService := services.NewFacturaAnulacionService(facturaAnulacionRepo, sucursalFacturadorRepo, logEnvioRepo, usuarioService)
	facturaAnulacionHandler := handlers.NewFacturaAnulacionHandler(facturaAnulacionService)

	envioWorker := services.NewEnvioWorker(facturaPrevaloradaService, facturaAnulacionService)
	go envioWorker.Iniciar()

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

	SetupRoutes(app, authHandler, usuarioService, dbConnectionHandler, consultasHandler, codigoProductoHandler, conexionSucursalHandler, usuarioHandler, sucursalFacturadorHandler, facturaPrevaloradaHandler, facturaAnulacionHandler, logEnvioHandler, resumenContableHandler)

	port := ":" + config.ServerPort
	log.Printf("Servidor ejecutándose en puerto %s", config.ServerPort)
	log.Printf("Endpoints disponibles:")
	log.Printf("  - Health Check: http://localhost%s/api/v1/health", port)
	log.Printf("  - Connections: http://localhost%s/api/v1/connections", port)

	if err := app.Listen(port); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
