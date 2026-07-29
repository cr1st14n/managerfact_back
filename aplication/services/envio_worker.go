package services

import (
	"log"
	"managerfact/internal/domain/models"
	"time"
)

// EnvioWorker envía automáticamente, en background, las facturas
// prevaloradas y las solicitudes de anulación que están en estado
// "pendiente" — ver doc/EnvioFacturacion.md sección 5.
//
// Aplica un circuit breaker por sucursal facturador: si una sucursal falla
// por un error de transporte (el facturador no responde / está caído), se
// la marca "en_revision" y se dejan de intentar el resto de sus pendientes
// en este ciclo — evita golpear un servidor caído con decenas de intentos
// seguidos. En el siguiente ciclo, si ya pasó el tiempo de espera
// (cooldownRevision), se vuelve a intentar; si el facturador responde
// (aceptado o rechazado, no importa cuál) la sucursal vuelve sola a
// "activo" — esa parte la hacen Facturar/Anular.
type EnvioWorker struct {
	facturaPrevalorada *FacturaPrevaloradaService
	facturaAnulacion   *FacturaAnulacionService
	intervalo          time.Duration
	cooldownRevision   time.Duration
	detener            chan struct{}
}

func NewEnvioWorker(facturaPrevalorada *FacturaPrevaloradaService, facturaAnulacion *FacturaAnulacionService) *EnvioWorker {
	return &EnvioWorker{
		facturaPrevalorada: facturaPrevalorada,
		facturaAnulacion:   facturaAnulacion,
		intervalo:          30 * time.Second,
		cooldownRevision:   5 * time.Minute,
		detener:            make(chan struct{}),
	}
}

// Iniciar corre el loop de envío; se llama con "go worker.Iniciar()".
func (w *EnvioWorker) Iniciar() {
	log.Printf("[EnvioWorker] iniciado (intervalo=%s, cooldown_revision=%s)", w.intervalo, w.cooldownRevision)
	ticker := time.NewTicker(w.intervalo)
	defer ticker.Stop()
	for {
		select {
		case <-w.detener:
			return
		case <-ticker.C:
			w.procesarPrevaloradas()
			w.procesarAnulaciones()
		}
	}
}

// Detener corta el loop; no hace falta esperar a que termine el ciclo en
// curso, es solo para tests/apagado ordenado.
func (w *EnvioWorker) Detener() {
	close(w.detener)
}

// sucursalDisponible decide si vale la pena intentar enviar hacia esta
// sucursal ahora mismo: no está en la lista de las que ya fallaron en este
// mismo ciclo, y si está "en_revision" de un ciclo anterior, ya pasó el
// tiempo de espera desde el último error.
func (w *EnvioWorker) sucursalDisponible(sucursal *models.SucursalFacturador, fallidasEnEsteCiclo map[uint]bool) bool {
	if fallidasEnEsteCiclo[sucursal.ID] {
		return false
	}
	if sucursal.EstadoConexion != "en_revision" {
		return true
	}
	if sucursal.UltimoErrorConexion == nil {
		return true
	}
	return time.Since(*sucursal.UltimoErrorConexion) >= w.cooldownRevision
}

func (w *EnvioWorker) procesarPrevaloradas() {
	pendientes, err := w.facturaPrevalorada.ListarPendientesParaEnvio()
	if err != nil {
		log.Printf("[EnvioWorker] error listando facturas prevaloradas pendientes: %v", err)
		return
	}

	fallidas := map[uint]bool{}
	for _, factura := range pendientes {
		if factura.SucursalFacturador == nil {
			continue
		}
		if !w.sucursalDisponible(factura.SucursalFacturador, fallidas) {
			continue
		}
		if _, err := w.facturaPrevalorada.Facturar(factura.ID, "automatico"); err != nil {
			log.Printf("[EnvioWorker] error facturando prevalorada id=%d: %v", factura.ID, err)
			fallidas[factura.SucursalFacturadorID] = true
		}
	}
}

func (w *EnvioWorker) procesarAnulaciones() {
	pendientes, err := w.facturaAnulacion.ListarPendientesParaEnvio()
	if err != nil {
		log.Printf("[EnvioWorker] error listando facturas de anulación pendientes: %v", err)
		return
	}

	fallidas := map[uint]bool{}
	for _, factura := range pendientes {
		if factura.SucursalFacturador == nil {
			continue
		}
		if !w.sucursalDisponible(factura.SucursalFacturador, fallidas) {
			continue
		}
		if _, err := w.facturaAnulacion.Anular(factura.ID, "automatico"); err != nil {
			log.Printf("[EnvioWorker] error anulando id=%d: %v", factura.ID, err)
			fallidas[factura.SucursalFacturadorID] = true
		}
	}
}
