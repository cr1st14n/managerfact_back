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

## 2. Registro de boletos y facturación prevalorada

Dos etapas independientes:
1. **Importación**: se cargan los datos base de cada boleto (Excel/lista), sin tocar todavía nada del formato que espera el facturador.
2. **Facturación**: el sistema arma el JSON (`datosGenerales` + `documentoFiscal`) combinando el boleto importado + la configuración de la sucursal facturador + una serie de valores fijos, lo envía al facturador, y guarda la respuesta (estado + detalle del evento).

Sigue siendo **un solo ítem** por factura (uso de derecho aeroportuario), modelado plano en una sola tabla, sin tabla hija de detalle.

### Etapa 1: Importación

Antes de importar el archivo se selecciona **una sola vez, para todo el lote**: a qué `sucursal_facturador` se enviarán estas facturas. Esa selección queda fija para todas las filas del archivo — `sucursal_facturador_id` no es una columna del Excel.

### Modelo: `facturas_prevaloradas`

| Campo | Origen |
|---|---|
| `sucursal_facturador_id` | seleccionado antes de importar, fijo para todo el lote |
| `lote_id` | UUID del lote de importación (ver sección 3) |
| `codigo_integracion` | generado (UUID), único — idempotency key |
| `tipo` | fijo `"FACTURA_PREVALORADA"` (no viene del Excel) |
| `observacion` | motivo de carga del lote, ingresado antes de importar, fijo para todo el lote |
| `detalle` | Excel — descripción del boleto/ítem |
| `codigo_producto` | Excel |
| `costo_dua_dolares` | Excel — costo del DUA, en dólares |
| `fecha_compra_boleto` | Excel — fecha de compra del boleto aéreo |
| `tipo_cambio` | Excel — tc vigente **a la fecha de `fecha_compra_boleto`**. Se guarda tal cual, sin recalcular |
| `fecha_emision` | Excel |

Seguimiento de envío (se completa en la etapa 2):

| Campo | Notas |
|---|---|
| `estado` | `pendiente` → `enviado` → `aceptado` / `rechazado` / `error` |
| `codigo_respuesta`, `mensaje_respuesta` | detalle del evento devuelto por el facturador (formato exacto: pendiente de definir) |
| `fecha_envio`, `fecha_respuesta` | |
| `intentos_consulta` | contador para el polling de estado |

### Etapa 2: Facturación (armado del JSON)

Los siguientes valores **no se guardan como columnas**: son constantes que el `FacturadorClient` inyecta directamente al armar el JSON de envío (`recibir-sincrono`), combinándolas con los campos de la etapa 1 y con los datos de `sucursales_facturador` (nit, código sucursal SIN, punto de venta, moneda BOB):

```
codigoCliente = "N/A"
tipoDocumentoIdentidad = "NIT"
numeroDocumento = "0"
nombreRazonSocial = "S/N"
complemento = null
celularCliente = null
tipoDocumentoFiscal = 1
tipoDocumentoSector = "23"
tipoEmision = 1
codigoExcepcion = null
metodoPago = "1"
codigoMoneda = sucursal.codigo_moneda_bob (ej. "BOB")
cantidad = 1
codigoUnidadMedida = "58"
codigoDetalleTransaccion = 1
montoDescuentoDetalle = null
```

> Nota (confirmado probando contra el servidor real de FacturaClic en
> `facturacion.lp.naabol.gob.bo`): `tipoDocumentoIdentidad` debe ser el
> string `"NIT"`, no `"5"` como decía el ejemplo original — con `"5"` el
> facturador responde `El campo 'tipoDocumentoIdentidad' 5 no se encontró
> registrado, usa NIT`.

Mapeo del boleto importado al payload del facturador:
- `descripcion` (detalle del ítem) = `detalle`
- `codigoProducto` = `codigo_producto`
- `precioUnitario` / `subtotal` / `montoTotal` / `montoTotalSujetoIva` = `costo_dua_dolares`
- `tipoCambio` = `tipo_cambio`
- `fechaEmision` = `fecha_emision`

### Pendiente de confirmar (no bloquea el resto del diseño)
- `montoTotalMoneda`: ¿= `costo_dua_dolares / tipo_cambio`?
- `fechaEmisionFactura`: ¿= `fechaEmision`?
- `emailCliente`: ¿vacío o algún valor por defecto?
- `usuario`: ¿el usuario que importa el Excel, o un valor fijo de sistema?

---

## 3. Importación desde Excel

**Endpoint**: `POST /api/v1/facturas-prevaloradas/importar-excel`
- Multipart: archivo `.xlsx` + `sucursal_facturador_id` + `observacion` (una sola sucursal y observación por archivo, fijas para todo el lote).
- Librería: `github.com/xuri/excelize/v2` (agregar a `go.mod`).

**Columnas esperadas** (por nombre de encabezado): `detalle`, `costo_dua_dolares`, `fecha_emision`, `fecha_compra_boleto`, `tipo_cambio` (tc a la fecha de `fecha_compra_boleto`), `codigo_producto`.

`tipo` **no** es columna del Excel: se fija en `"FACTURA_PREVALORADA"` al guardar cada fila, ya que este importador solo maneja prevaloradas por ahora.

**Por fila**:
1. Validar tipos y campos requeridos. Fila inválida → no se guarda, se reporta el motivo; no aborta el archivo completo.
2. Generar `codigo_integracion` (UUID).
3. Completar el resto del payload con los defaults fijos de la sección 2.
4. Guardar con `estado = "pendiente"` y el `lote_id` del archivo (UUID generado al iniciar la importación; no se crea una tabla `lotes_importacion` aparte — el progreso se calcula agregando sobre `facturas_prevaloradas WHERE lote_id = ?`).

**Respuesta**: `lote_id`, total de filas, válidas, con error (detalle fila + motivo).

### Endpoints de seguimiento
- `GET /api/v1/facturas-prevaloradas/plantilla` — descarga el `.xlsx` de ejemplo con las columnas esperadas.
- `GET /api/v1/facturas-prevaloradas/lotes` — registro de lotes: sucursal facturador, tipo, total y desglose por estado de cada lote importado.
- `GET /api/v1/facturas-prevaloradas?estado=&lote_id=` — detalle de un lote (o de todas las facturas, filtrando por estado).
- `GET /api/v1/facturas-prevaloradas/:id`

---

## 4. Anulación de facturas

Segundo tipo de lote (aparte de la prevalorada), con su propia plantilla de Excel. No describe un ítem nuevo a facturar — solo referencia una factura ya emitida (por CUF + código de integración) y el motivo de anulación — por eso vive en su propia tabla (`facturas_anulacion`) en vez de columnas nullable sobre `facturas_prevaloradas`.

### Modelo: `facturas_anulacion`

| Campo | Origen |
|---|---|
| `sucursal_facturador_id` | seleccionado antes de importar, fijo para todo el lote |
| `lote_id` | UUID del lote de importación |
| `observacion` | motivo de carga del lote, fijo para todo el lote (igual que en prevaloradas) |
| `codigo_integracion` | **Excel** — a diferencia de la prevalorada, acá NO se genera: es el código de integración de la factura original que se quiere anular |
| `cuf` | Excel — CUF de la factura original a anular |
| `codigo_motivo` | Excel — código de motivo de anulación |

Seguimiento de envío (mismos campos que `facturas_prevaloradas`): `estado`, `codigo_respuesta`, `mensaje_respuesta`, `fecha_envio`, `fecha_respuesta`, `intentos_consulta`.

### Importación desde Excel

**Endpoint**: `POST /api/v1/facturas-anulacion/importar-excel`
- Multipart: archivo `.xlsx` + `sucursal_facturador_id` + `observacion` (una sola sucursal y observación por archivo, fijas para todo el lote).
- **Columnas esperadas**: `cuf`, `codigo_motivo`, `codigo_integracion` (de la factura original a anular).
- Mismas reglas de importación por fila que la prevalorada: fila inválida → no se guarda, se reporta el motivo; no aborta el archivo completo.

### Endpoints de seguimiento
- `GET /api/v1/facturas-anulacion/plantilla` — descarga el `.xlsx` de ejemplo con las columnas esperadas.
- `GET /api/v1/facturas-anulacion/lotes` — registro de lotes: sucursal facturador, observación, total y desglose por estado de cada lote importado.
- `GET /api/v1/facturas-anulacion?estado=&lote_id=` — detalle de un lote (o de todas las facturas, filtrando por estado).
- `GET /api/v1/facturas-anulacion/:id`

### Pendiente de confirmar (no bloquea el resto del diseño)
- Forma exacta del payload de envío al facturador para anulación (endpoint distinto de `recibir-sincrono`, probablemente algo como `anular-sincrono` — a confirmar contra la documentación de FacturaClic).

---

## 5. Envío automático

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
------------------------------------------------------- PRUEBA JSON testeado
para envio de prevalorada clic-core/facturas/recibir-sincrono
para modificar la fecha de emision seria asi
$formattedDateTime = preg_replace('/(\.\d{3})\d+/', '$1', Carbon::now('America/La_Paz')->subMinute()->format('Y-m-d\TH:i:s.uP'));
"tipoEmision" => 3,
"fechaEmision" => $formattedDateTime,




{
    "datosGenerales": {
        "nitEmisor": "419945029",
        "sucursalEmisor": 0,// proviene de la sucursal
        "puntoVentaEmisor": "",
        "codigoIntegracion": "KKKKKKKK", la uuid 
        "codigoCliente": "N/A",
        "celularCliente": null,
        "emailCliente": null,
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
            "tipoDocumentoIdentidad": "NIT",
            "numeroDocumento": "0",
            "complemento": null,
            "fechaEmisionFactura": null,
            "metodoPago": "1",
            "codigoMoneda": "BOB",
            "tipoCambio": 1,
            "montoTotalMoneda": 10,
            "montoTotal": 10,
            "montoTotalSujetoIva": 10,
            "usuario": "Client-Postman"
        },
        "detalle": [
            {
                "codigoProducto": "99101",
                "descripcion": "PIZARRA BITUMINOSA O DE ACEITE Y ARENAS DE ALQUITRÁN",
                "cantidad": 1,
                "precioUnitario": 10,
                "subtotal": 10,
                "montoDescuentoDetalle": null,
                "codigoDetalleTransaccion": 1,
                "codigoUnidadMedida": "58"
            }
        ]
    }
}
respuesta succes
{
    "codigo": 200,
    "respuesta": "OK",
    "mensaje": "Documento fiscal procesado de manera correcta",
    "urlDocumento": "https://facturacion.lp.naabol.gob.bo:8443/clic-portal/df/5239ddebb98ca5ff80744b3be577a064",
    "idDocumento": 7737942,
    "tipoEmision": 1,
    "tipoEmisionDescripcion": "ONLINE",
    "cuf": "1CBBDAA372E42FF1FE90B7CD09E4B0B4A88660F66512F40053660BF74",
    "cufd": "BQXxDdMONcEFBN00FFRjBCNUM5MUU=Qj4lVUtVZEhhVUFMyMjRFNTcxQ0ZGQ",
    "cuis": "EC9202A9",
    "numeroFactura": 17735,
    "fechaEmision": "2026-07-29T01:17:07.584-04:00",
    "estadoDocumentoFiscal": "VERIFICADO",
    "codigoRecepcionSin": "beabe2d6-8b0c-11f1-9c04-abe440964172",
    "codigoIntegracion": "KKKKKKKK",
    "urlSin": "https://siat.impuestos.gob.bo/consulta/QR?nit=419945029&cuf=1CBBDAA372E42FF1FE90B7CD09E4B0B4A88660F66512F40053660BF74&numero=17735&t=2",
    "leyenda": "Ley N° 453: El proveedor de servicios debe habilitar medios e instrumentos para efectuar consultas y reclamaciones."
}
si falla algo
{
    "codigo": 400,
    "respuesta": "NOK",
    "mensaje": "El campo 'tipoDocumentoIdentidad' debe ser: 5",
    "urlDocumento": null,
    "idDocumento": 0,
    "tipoEmision": 0,
    "tipoEmisionDescripcion": null,
    "cuf": null,
    "cufd": null,
    "cuis": null,
    "numeroFactura": null,
    "fechaEmision": null,
    "estadoDocumentoFiscal": null,
    "codigoRecepcionSin": null,
    "codigoIntegracion": "KKKKKKKK",
    "urlSin": null,
    "leyenda": null
}
---
para anular  clic-core/facturas/anular
{
    "datosGenerales": {
        "nitEmisor": 419945029,
        "sucursalEmisor": "0",
        "puntoVentaEmisor": "",
        "canalFacturacion": "core"
    },
    "documentoFiscal": {
        "codigoIntegracion": "KKKKKKKK2",
        "cuf": "1CBBDAA372E42FF1FE90B7CD09E4B0B4A88660F66512F40053660BF74",
        "codigoMotivo": "1"
    }
}
respuesta todo OK
{
    "codigo": 200,
    "respuesta": "OK",
    "mensaje": "Documento Fiscal con CUF:1CBBDAA372E42FF1FE90B7CD09E4B0B4A88660F66512F40053660BF74 ANULADO de manera exitosa.",
    "codigoIntegracion": "KKKKKKKK"
}
si falla algo
{
    "codigo": 400,
    "respuesta": "NOK",
    "mensaje": "El documento fiscal con CUF[1CBBDAA372E42FF1FE90B7CD09E4B0B4A88660F66512F40053660BF74] ya fue ANULADO",
    "codigoIntegracion": "KKKKKKKK2"
}