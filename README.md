# prj_withdrawal · Cajero Automático (Go)

![CI](https://github.com/USUARIO/prj_withdrawal/actions/workflows/ci.yml/badge.svg)

Proyecto del curso **DSL / Gherkin-BDD** (UNI-FIIS). Implementa los casos de uso
**Retiro de efectivo**, **Abono en efectivo** e **Integridad de la transacción**
con desarrollo iterativo e incremental (BDD + TDD) y una **App WEB** final.

| Iteración | Alcance | Versión |
|-----------|---------|---------|
| 1 | codeLAB Semana-7c (punto de partida) + US-01b Abono | v1.3.0 |
| 2 | US-04 Integridad de transacción + log JSON + CLI interactivo | v2.0.0 |
| 3 | Refactorización por capas + App WEB (esta versión) | v3.0.0 |

## Ejecutar

Requiere Go 1.22+.

```bash
# App WEB  ->  http://127.0.0.1:8080
go run ./cmd/atm-web

# CLI interactivo
go run ./cmd/atm-cli

# CLI no interactivo (imprime la transacción en JSON)
go run ./cmd/atm-cli -op retiro -nombre Alice -monto 200
go run ./cmd/atm-cli -historial -nombre Alice
```

La primera ejecución copia `data/cuentas.seed.json` a `data/cuentas.json`
(fuente de datos) y crea `data/transacciones.json` (log de auditoría).

Con Docker: `docker build -t atm . && docker run -p 8080:8080 atm`

## Pruebas

```bash
go test -race -cover ./...
```

Los escenarios de `features/*.feature` están traducidos a pruebas con la misma
estructura Given/When/Then en `internal/servicio/bdd_test.go`.

## Arquitectura

```
cmd/atm-web        App WEB (servidor HTTP)
cmd/atm-cli        CLI interactivo / por flags
internal/dominio   Entidades y reglas puras: Dinero (centavos), Cuenta, Transaccion
internal/almacen   Interfaz de persistencia + implementación JSON (escritura atómica)
internal/servicio  Casos de uso: atomicidad (RN-03) + auditoría (US-04), thread-safe
internal/api       API REST + interfaz web embebida (go:embed)
internal/cli       Menú interactivo
features/          Especificación Gherkin (retiro, abono, integridad)
```

Decisiones de diseño:

- **Dinero como `int64` en centavos** (Modelo de Dominio): evita errores de
  redondeo de `float64`.
- **Atomicidad**: el saldo se guarda y se audita; si algo falla, se revierte en
  memoria y en disco. Todo intento (éxito, rechazo o error) deja rastro.
- **Dependencias invertidas**: `servicio` depende de la interfaz `almacen.Almacen`.

## API REST

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET  | `/api/cuentas` | Lista de cuentas y saldos |
| GET  | `/api/cuentas/{nombre}` | Saldo de una cuenta |
| POST | `/api/operaciones` | `{"cuenta":"Alice","tipo":"retiro","monto":"200.00"}` |
| GET  | `/api/transacciones?cuenta=Alice` | Historial (log JSON) |
| GET  | `/healthz` | Estado del servicio |

Códigos de `POST /api/operaciones`: `200` EXITOSA · `422` RECHAZADA (saldo
insuficiente) · `400` ERROR de validación · `404` cuenta inexistente · `500` fallo de persistencia.

## Publicar en GitHub

```bash
git init && git add . && git commit -m "feat: prj_withdrawal v3.0.0"
git branch -M main
git remote add origin https://github.com/USUARIO/prj_withdrawal.git
git push -u origin main
```

## Limitaciones conocidas

- El log se reescribe completo en cada operación (suficiente para el alcance
  académico; para volumen real usar JSON Lines o una base de datos).
- No hay autenticación: el modelo asume que el cliente ya está autenticado (precondición del caso de uso).

## Licencia

MIT
