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
// (transcrito de doc/sucursales.md). Corre en cada arranque y usa
// FirstOrCreate por cada regional (por nombre, único) y cada sucursal (por
// codigo_sucursal_sin, único) en vez de saltarse todo si ya hay datos: así,
// si se agrega una sucursal nueva a este catálogo más adelante (como pasó
// con el código 0 "Oficina Central"), el próximo reinicio la crea sola sin
// duplicar ni tocar las que ya existen. Pendiente de validar contra
// SFE_SUCURSAL en producción, tal como advierte el propio documento de
// origen.
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
		if err := db.Where(models.Regional{Nombre: nombreRegional}).FirstOrCreate(&regional).Error; err != nil {
			return fmt.Errorf("error creando regional %s: %v", nombreRegional, err)
		}
		for _, s := range sucursalesPorRegional[nombreRegional] {
			sucursal := models.SucursalCatalogo{
				CodigoSucursalSin: s.Codigo,
				Nombre:            s.Nombre,
				RegionalID:        regional.ID,
			}
			// Condición como string+args (no struct): un Where con struct
			// omite los campos en su valor cero, y codigo_sucursal_sin=0
			// ("Oficina Central") es justamente ese caso — con struct nunca
			// se filtraría por código y quedaría sin sembrar.
			if err := db.Where("codigo_sucursal_sin = ?", s.Codigo).FirstOrCreate(&sucursal).Error; err != nil {
				return fmt.Errorf("error creando sucursal %s: %v", s.Nombre, err)
			}
		}
	}

	log.Println("Catálogo de regionales y sucursales verificado/completado")
	return nil
}

// SetupRoutes configura todas las rutas de la API
func SetupRoutes(
	app *fiber.App,
	authHandler *handlers.AuthHandler,
	usuarioService *services.UsuarioService,
	dbConnectionHandler *handlers.DbConnectionHandler,
	consultasHandler *handlers.ConsultasHandler,
	codigoProductoHandler *handlers.CodigoProductoHandler,
	usuarioHandler *handlers.UsuarioHandler,
	sucursalFacturadorHandler *handlers.SucursalFacturadorHandler,
	facturaPrevaloradaHandler *handlers.FacturaPrevaloradaHandler,
	facturaAnulacionHandler *handlers.FacturaAnulacionHandler,
	logEnvioHandler *handlers.LogEnvioHandler,
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

	// Ruta de health check (pública)
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "API funcionando correctamente",
		})
	})

	// Login (público)
	authHandler.RegisterRoutes(api)

	// Todo lo demás requiere sesión (JWT) — ver
	// infraestructura/middleware/auth_middleware.go
	protegido := api.Group("/", middleware.RequireAuth())

	// Usuarios y Sucursales Facturador además requieren rol "admin": los
	// operadores no pueden ver ni operar estos dos módulos. El middleware se
	// pasa directo a cada RegisterRoutes para que quede scopeado a sus
	// propios prefijos (/usuarios, /sucursales-facturador, etc.) — un grupo
	// con prefijo "/" aplicaría el middleware a TODO el resto de rutas
	// protegidas, no solo a estas dos.
	requireAdmin := middleware.RequireAdmin(usuarioService)

	// Registrar rutas de conexiones de BD
	dbConnectionHandler.RegisterRoutes(protegido)
	// Registrar rutas de consultas
	consultasHandler.RegisterRoutes(protegido)
	// Registrar rutas de codigo producto
	codigoProductoHandler.RegisterRoutes(protegido)
	// Registrar rutas de usuarios/regionales/catálogo de sucursales (solo admin)
	usuarioHandler.RegisterRoutes(protegido, requireAdmin)
	// Registrar rutas de sucursales facturador (FacturaClic) (solo admin)
	sucursalFacturadorHandler.RegisterRoutes(protegido, requireAdmin)
	// Registrar rutas de facturas prevaloradas (boletos)
	facturaPrevaloradaHandler.RegisterRoutes(protegido)
	// Registrar rutas de facturas de anulación
	facturaAnulacionHandler.RegisterRoutes(protegido)
	// Registrar rutas de logs de envío
	logEnvioHandler.RegisterRoutes(protegido)
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

	// usuarios (antes de consultas: ConsultasHandler necesita usuarioService
	// para verificar accesos por sucursal)
	usuarioRepo := repositories.NewUsuarioRepository(db)
	usuarioService := services.NewUsuarioService(usuarioRepo)
	usuarioHandler := handlers.NewUsuarioHandler(usuarioService)

	// login / sesión
	authHandler := handlers.NewAuthHandler(usuarioService)

	// Iniciar consultas
	consultasRepositori := repositories.NewConsutasRepository(db)
	consultaHandler := services.NewConsultasService(consultasRepositori)
	consultasHandler := handlers.NewConsultasHandler(consultaHandler, usuarioService)

	// codigo producto
	codigoProductoRepo := repositories.NewCodigoProductoRepoRepo(db)
	codigoProductoService := services.NewCodigoProductoService(codigoProductoRepo)
	codigoProductoHandler := handlers.NewCodigoProductoHandler(codigoProductoService)

	// sucursales facturador (FacturaClic)
	sucursalFacturadorRepo := repositories.NewSucursalFacturadorRepository(db)
	sucursalFacturadorService := services.NewSucursalFacturadorService(sucursalFacturadorRepo)
	sucursalFacturadorHandler := handlers.NewSucursalFacturadorHandler(sucursalFacturadorService)

	// logs de envío (registro de intentos, manuales y automáticos)
	logEnvioRepo := repositories.NewLogEnvioRepository(db)
	logEnvioHandler := handlers.NewLogEnvioHandler(logEnvioRepo)

	// facturas prevaloradas (boletos)
	facturaPrevaloradaRepo := repositories.NewFacturaPrevaloradaRepository(db)
	facturaPrevaloradaService := services.NewFacturaPrevaloradaService(facturaPrevaloradaRepo, sucursalFacturadorRepo, logEnvioRepo, usuarioService)
	facturaPrevaloradaHandler := handlers.NewFacturaPrevaloradaHandler(facturaPrevaloradaService)

	// facturas de anulación
	facturaAnulacionRepo := repositories.NewFacturaAnulacionRepository(db)
	facturaAnulacionService := services.NewFacturaAnulacionService(facturaAnulacionRepo, sucursalFacturadorRepo, logEnvioRepo, usuarioService)
	facturaAnulacionHandler := handlers.NewFacturaAnulacionHandler(facturaAnulacionService)

	// envío automático de pendientes (prevaloradas + anulación) en background
	envioWorker := services.NewEnvioWorker(facturaPrevaloradaService, facturaAnulacionService)
	go envioWorker.Iniciar()

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
	SetupRoutes(app, authHandler, usuarioService, dbConnectionHandler, consultasHandler, codigoProductoHandler, usuarioHandler, sucursalFacturadorHandler, facturaPrevaloradaHandler, facturaAnulacionHandler, logEnvioHandler)

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
