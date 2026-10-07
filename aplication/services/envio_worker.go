package services

import (
	"log"
	"managerfact/internal/domain/models"
	"time"
)

// Un error de transporte pausa la sucursal hasta cooldownRevision para no saturar un servidor caído.
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

func (w *EnvioWorker) Detener() {
	close(w.detener)
}

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
