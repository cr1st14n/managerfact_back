# Módulo: Envío de Facturación (FacturaClic)

> Estado: planificación. Notas originales del requerimiento conservadas en la sección final.

## Flujo general

1. Se registra cada **sucursal** con sus datos de conexión al facturador (FacturaClic).
2. Llegan **listas de facturas prevaloradas** (Excel) que se **registran** en el sistema.
3. Un proceso automático las va **enviando una por una** al facturador de la sucursal correspondiente.
4. Si el facturador responde rápido, se guarda el resultado en el momento. Si demora, la factura queda `pendiente` y se consulta su estado más tarde (FacturaClic expone un servicio de consulta de estado).

---

## 1. Registro de sucursal

Tabla nueva e independiente del catálogo maestro `sucursales_catalogo` (que ya existe para accesos de usuario) — esta es la configuración de conexión al facturador por sucursal.

### Modelo: `sucursales_facturador`

| Campo | Tipo | Notas |
|---|---|---|
| `id` | uint | PK |
| `nombre` | string | |
| `codigo_sucursal_sin` | int | Código SIN de la sucursal (`sucursalEmisor` en el payload) |
| `punto_venta_emisor` | string | `puntoVentaEmisor` en el payload; puede ir vacío |
| `url_link_facturador` | string | Base URL del servidor FacturaClic de esta sucursal |
| `token_acceso` | string | **Cifrado en BD** (AES), nunca se expone completo por la API |
| `codigo_moneda_bob` | string | |
| `codigo_ci` | string | |
| `codigo_nit` | string | `nitEmisor` en el payload |
| `activo` | bool | |
| `created_at` / `updated_at` / `deleted_at` | timestamps | soft delete |

### Reglas
- El token se cifra al guardar (clave desde variable de entorno, ej. `FACTURADOR_TOKEN_KEY`) y se descifra solo en memoria al momento de armar la request HTTP saliente.
- Las respuestas de la API (`GET`/`list`) **nunca** devuelven el token en texto plano — a lo sumo un booleano `token_configurado` o los últimos 4 caracteres.

### Endpoints
- `POST /api/v1/sucursales-facturador`
- `GET /api/v1/sucursales-facturador`
- `GET /api/v1/sucursales-facturador/:id`
- `PUT /api/v1/sucursales-facturador/:id`
- `DELETE /api/v1/sucursales-facturador/:id` (soft delete)

### Capas (siguiendo el patrón existente `models → repositories → services → handlers`)
- `models.SucursalFacturador`
- `repositories.SucursalFacturadorRepository` (CRUD)
- `services.SucursalFacturadorService` (CRUD + cifrado/descifrado del token)
- `handlers.SucursalFacturadorHandler`

---

## 2. Registro de factura prevalorada

Siempre trae **un solo ítem** (uso de derecho aeroportuario), así que se modela plano en una sola tabla, sin tabla hija de detalle.

### Modelo: `facturas_prevaloradas`

Cabecera (mapea `datosGenerales` + `documentoFiscal.cabecera` del payload de FacturaClic):

| Campo | Origen |
|---|---|
| `sucursal_facturador_id` | FK a `sucursales_facturador` |
| `codigo_integracion` | generado (UUID), único — idempotency key |
| `lote_id` | UUID del lote de importación (ver sección 3) |
| `codigo_cliente`, `celular_cliente`, `email_cliente` | Excel / defaults |
| `tipo_documento_fiscal`, `tipo_documento_sector`, `codigo_excepcion`, `tipo_emision` | defaults fijos |
| `fecha_emision`, `fecha_emision_factura` | Excel |
| `nombre_razon_social`, `tipo_documento_identidad`, `numero_documento`, `complemento` | defaults fijos |
| `metodo_pago`, `codigo_moneda` | defaults fijos |
| `tipo_cambio`, `monto_total_moneda`, `monto_total`, `monto_total_sujeto_iva` | Excel / calculado |
| `usuario` | quién importa / valor fijo (a confirmar) |

Ítem (aplanado, mapea `documentoFiscal.detalle[0]`):

| Campo | Origen |
|---|---|
| `codigo_producto` | Excel |
| `descripcion_item` | Excel |
| `cantidad` | default `1` |
| `precio_unitario`, `subtotal` | = `monto` del Excel |
| `monto_descuento_detalle` | default `null` |
| `codigo_detalle_transaccion` | default `1` |
| `codigo_unidad_medida` | default `"58"` |

Seguimiento de envío:

| Campo | Notas |
|---|---|
| `estado` | `pendiente` → `enviado` → `aceptado` / `rechazado` / `error` |
| `codigo_respuesta`, `mensaje_respuesta` | de la respuesta de FacturaClic (formato exacto: pendiente de definir) |
| `fecha_envio`, `fecha_respuesta` | |
| `intentos_consulta` | contador para el polling de estado |

`tipo` (string, default `"prevalorada"`) para acomodar otros tipos de factura que el doc original menciona a futuro, sin construir hoy una abstracción genérica.

### Defaults fijos confirmados para "derecho aeroportuario"
```
codigoCliente = "N/A"
tipoDocumentoIdentidad = "5"
numeroDocumento = "0"
nombreRazonSocial = "S/N"
complemento = null
celularCliente = null
tipoDocumentoFiscal = 1
tipoDocumentoSector = "23"
tipoEmision = 1
codigoExcepcion = null
metodoPago = "1"
codigoMoneda = "1"
cantidad = 1
codigoUnidadMedida = "58"
codigoDetalleTransaccion = 1
montoDescuentoDetalle = null
```

### Pendiente de confirmar (no bloquea el resto del diseño)
- `montoTotalMoneda`: ¿= `monto / tipoCambio`?
- `fechaEmisionFactura`: ¿= `fechaEmision`?
- `emailCliente`: ¿vacío o algún valor por defecto?
- `usuario`: ¿el usuario que importa el Excel, o un valor fijo de sistema?

---

## 3. Importación desde Excel

**Endpoint**: `POST /api/v1/facturas-prevaloradas/importar-excel`
- Multipart: archivo `.xlsx` + `sucursal_facturador_id` (una sola sucursal por archivo).
- Librería: `github.com/xuri/excelize/v2` (agregar a `go.mod`).

**Columnas esperadas** (por nombre de encabezado): `fecha_emision`, `monto`, `tipo_cambio`, `descripcion`, `codigo_producto`.

**Por fila**:
1. Validar tipos y campos requeridos. Fila inválida → no se guarda, se reporta el motivo; no aborta el archivo completo.
2. Generar `codigo_integracion` (UUID).
3. Completar el resto del payload con los defaults fijos de la sección 2.
4. Guardar con `estado = "pendiente"` y el `lote_id` del archivo (UUID generado al iniciar la importación; no se crea una tabla `lotes_importacion` aparte — el progreso se calcula agregando sobre `facturas_prevaloradas WHERE lote_id = ?`).

**Respuesta**: `lote_id`, total de filas, válidas, con error (detalle fila + motivo).

### Endpoints de seguimiento
- `GET /api/v1/facturas-prevaloradas?estado=&lote_id=`
- `GET /api/v1/facturas-prevaloradas/:id`

---

## 4. Envío automático

- `FacturadorClient`: cliente HTTP que arma el JSON anidado (`datosGenerales` / `documentoFiscal`) a partir de la sucursal + la factura, y llama:
  - `POST {url_link_facturador}/clic-core/facturas/recibir-sincrono`
  - Header `Authorization: Bearer {token_acceso}` (descifrado en memoria)
- `EnvioWorker`: goroutine/ticker en background que toma facturas en `pendiente` (ordenadas por `lote_id` + orden de creación) y las envía una por una.
  - Respuesta rápida → guarda `codigo_respuesta` / `estado` / `fecha_respuesta` de inmediato.
  - Timeout / sin respuesta → la deja en `pendiente` para reintentar consulta de estado después (endpoint de consulta: **pendiente de definir**).
- `POST /api/v1/facturas-prevaloradas/:id/consultar-estado` — disparo manual, además del polling automático.

### Bloqueos pendientes para cerrar el cliente HTTP
- Forma exacta de la respuesta de `recibir-sincrono` (campos de código de respuesta / estado, cómo luce un rechazo vs. una aceptación).
- Endpoint y forma de la consulta de estado para las facturas que quedan `pendiente`.

---

## Notas originales del requerimiento (fuente)

```
- registro de sucursales
  Nombre, codigoSucursal SIN, UrlLinkFacturador, token de acceso, codigo moneda BOB, codigo CI, CODIGO Nit

- registro de facturas
 - factura prevalorada: fecha emision, monto, tipo cambio, descripcion, codigo de producto,
   codigo de integracion (uuid o random de 10 dígitos) | datos de respuesta: codigo respuesta, estado

 existiran otros tipos pero empezaremos por esta
```

Ejemplo de payload de envío (`recibir-sincrono`), factura real de un solo ítem (derecho aeroportuario):

```json
{
    "datosGenerales": {
        "nitEmisor": "419945029",
        "sucursalEmisor": 29,
        "puntoVentaEmisor": "",
        "codigoIntegracion": "1025000001832577",
        "codigoCliente": "N/A",
        "celularCliente": null,
        "emailCliente": "vlopez@mc4.com.boo",
        "atributosAdicionalesGeneral": []
    },
    "documentoFiscal": {
        "cabecera": {
            "tipoDocumentoFiscal": 1,
            "tipoDocumentoSector": "23",
            "codigoExcepcion": null,
            "tipoEmision": 1,
            "fechaEmision": null,
            "nombreRazonSocial": "S/N",
            "tipoDocumentoIdentidad": "5",
            "numeroDocumento": "0",
            "complemento": null,
            "fechaEmisionFactura": null,
            "metodoPago": "1",
            "codigoMoneda": "1",
            "tipoCambio": 6.96,
            "montoTotalMoneda": 2,
            "montoTotal": 13.9,
            "montoTotalSujetoIva": 13.9,
            "usuario": "Client-Postman"
        },
        "detalle": [
            {
                "codigoProducto": "99101",
                "descripcion": "PIZARRA BITUMINOSA O DE ACEITE Y ARENAS DE ALQUITRÁN",
                "cantidad": 1,
                "precioUnitario": 13.9,
                "subtotal": 13.9,
                "montoDescuentoDetalle": null,
                "codigoDetalleTransaccion": 1,
                "codigoUnidadMedida": "58"
            }
        ]
    }
}
```
