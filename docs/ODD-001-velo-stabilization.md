# ODD-001 — Estabilización de VELO (v0.1.0 → v0.2.0)

> Orden de Delegación Determinista. Cada Work Package (WP) es autocontenido, tiene archivos con dueño exclusivo y una Definition of Done ejecutable. Un worker **no** toca archivos fuera de su lista.

- **Fecha:** 2026-11-04 · **Base:** `main@8d8176e` · **Autoría:** Moisés Vera
- **Objetivo:** que `velo init` produzca un proyecto que **compila, despliega y es seguro**, y que `velo verify` acepte un Passport-Context real de OpenPassport.
- **Cambio de stack:** InsForge pasa a *legacy*. Default nuevo: **SurrealDB + Cloudflare Pages/Workers**. Frontends: Astro (default), Next.js, React-Vite.

---

## 1. Hallazgos de auditoría (estado real, verificado)

| # | Sev | Hallazgo | Evidencia |
|---|---|---|---|
| H1 | **CRÍTICO** | **No existe toolchain Go en la máquina.** Nada de VELO se compiló ni se probó localmente. | `go: command not found` (bash y pwsh). `go.mod` pide `go 1.27.1`. |
| H2 | **CRÍTICO** | **`velo verify` es incompatible con OpenPassport v1.** VELO valida un struct "Contract V0" (`project/stack/layers/sha256` en raíz) re-serializado por Go. El `manifest.json` real de OpenPassport usa `files[]` y el self-hash va en la entrada `manifest.json` con modo `raw-bytes-with-own-sha256-replaced-by-64-zeroes`. Resultado: todo pasaporte real **falla**. | `pkg/manifest/manifest.go:51-82` vs `openpassport/templates/open-passport-template/manifest.json`. Verificado: el template OP es consistente (Go raw-bytes = TS = `35a9fea4…9d92`, 0 archivos con mismatch). |
| H3 | **CRÍTICO (seg.)** | Webhook Telegram generado es **fail-open**: si `TELEGRAM_WEBHOOK_SECRET` no está, acepta todo. Usa `!==` (no constant-time). No valida `TELEGRAM_OWNER_ID`. No hay flujo Approve→commit. El README afirma lo contrario. | `internal/scaffold/scaffold.go` (bloque `telegramAPI`, `if (expectedSecret && …)`). README "4 Commandments" #3. |
| H4 | **ALTO (seg.)** | `InsForgeAnonKey` se escribe en `.env.example`, que **se commitea** → fuga de credenciales al repo del cliente. | `scaffold.go` bloque `envExample`. |
| H5 | ALTO | **Cero tests** en el repo. | `git ls-files` → ningún `*_test.go`. |
| H6 | ALTO | `--frontend react-vite` se acepta pero se ignora: siempre genera Astro. | `scaffold.go` no ramifica por `state.Frontend`. |
| H7 | ALTO | Ruta headless no sanitiza slug: `--name "Mi App"` crea carpeta con espacios. El wizard sí usa `slugify`. | `cmd/velo/main.go` (`finalSlug = *name`). |
| H8 | MEDIO | BaaS/Runtime/Cockpit hardcodeados en headless; no hay flag `--data`. | `cmd/velo/main.go` `handleInit`. |
| H9 | MEDIO | Sin detección de TTY: `velo` en CI/agent cuelga en el wizard. | `runWizard()` llamado sin chequeo. |
| H10 | MEDIO | Scaffold monolítico con strings inline (192 líneas de `fmt.Sprintf`) — no testeable ni extensible a providers. | `internal/scaffold/scaffold.go`. |
| H11 | MEDIO | `wrangler.toml` con `compatibility_date = "2024-11-01"`; directorio `src/lib/insforge/` vacío; sin `.dev.vars.example`. | `scaffold.go`. |
| H12 | BAJO | Inconsistencia de versión Go: `go.mod 1.27.1` / README badge `>=1.22` / CI `'>=1.22'`. Sin CI de PR (solo release por tag). | `.github/workflows/release.yml`, README. |
| H13 | BAJO | `wizard.Model.state` se crea y nunca se usa (los steps comparten otro puntero). Código muerto. | `internal/wizard/model.go:15-24`. |
| H14 | MEDIO (OP) | OpenPassport tiene **dos contratos**: `pkg/schema/manifest.go` (Contract V0, el que copió VELO) está muerto respecto al `PassportManifest` real que usan `core.go` y TS. | `openpassport/pkg/schema/manifest.go`. |

---

## 2. Contratos congelados (los workers NO los renegocian)

### C1 — Verificación de pasaporte (OpenPassport v1)
1. Leer `manifest.json` como **bytes crudos**.
2. Tomar `sha256` de la entrada `files[]` con `path == "manifest.json"` → `cur`.
3. Self-hash = `sha256( LF( replaceFirst(raw, cur, manifestSelfHash.placeholder) ) )`, donde `LF(s)` = `\r\n|\r → \n`, `trim`, + `"\n"`.
4. Cada otro `files[i]`: `sha256(LF(contenido)) == files[i].sha256`.
5. Hex minúscula. Fail-closed ante cualquier campo ausente.
- **Golden vector:** `openpassport/templates/open-passport-template/manifest.json` → `35a9fea4f93b0da51284c57f9dbd7dab7140dc4ab702176b9f23ca2269519d92`.

### C2 — `wizard.State` (campos nuevos, nombres exactos)
```go
Frontend     string // "astro" | "next" | "react-vite"
Runtime      string // "cloudflare-pages" | "cloudflare-workers"
DataProvider string // "surrealdb" (default) | "insforge" (legacy) | "none"
DataURL      string // URL pública del endpoint; NUNCA claves
Cockpit      string // "telegram" | "none"
```
`BaaS`, `InsForgeProjectID`, `InsForgeAnonKey` se **eliminan** (migrar a `DataProvider`/`DataURL`). Ninguna clave/secreto entra a `State` ni a archivos commiteables.

### C3 — Layout de templates
```
internal/scaffold/templates/
  base/            (velo.config.ts, .gitignore, .dev.vars.example, wrangler.toml)
  frontend/astro/  frontend/next/  frontend/react-vite/
  data/surrealdb/  data/insforge/  data/none/
  cockpit/telegram/
```
`embed.FS` + `text/template`. Archivo `*.tmpl` → se renderiza y pierde la extensión. Orden de composición: base → frontend → data → cockpit (último gana, conflicto = error).

### C4 — Cockpit Telegram (seguridad)
- Secreto ausente ⇒ **503 fail-closed**.
- Comparación constant-time (`crypto.subtle.timingSafeEqual` en Workers, sobre `TextEncoder` de igual longitud).
- `from.id` ≠ `TELEGRAM_OWNER_ID` ⇒ 403.
- `callback_query.data` con formato `approve:<id>` / `reject:<id>`; la mutación solo se ejecuta sobre una `pending_action` existente y no expirada.

---

## 3. Protocolo de worker (aplica a todos)

1. Rama `odd001/<WP-id>` desde `main` actualizado. Un WP = un PR.
2. **Prohibido:** tocar archivos fuera de tu lista; borrar/debilitar asserts; editar goldens sin `RATIONALE.md` en el PR; dependencias sin versión exacta; red en tests.
3. Antes de entregar: `gofmt -l . ` vacío, `go vet ./...`, `go test ./... -count=1`, y adjuntar `git diff --stat` en el PR.
4. Si un contrato (C1–C4) no alcanza: **parar y escalar**, no improvisar.
5. Docs externas (SurrealDB SDK, Astro, OpenNext, Wrangler): leer la versión pinneada antes de escribir código. No inventar APIs.

---

## 4. Work Packages y olas

```
Ola 0 (bloqueante, humano)   W0
Ola 1 (paralelo)             W1 · W3 · W6 · W8
Ola 2 (paralelo)             W2a · W4
Ola 3 (paralelo)             W2b · W2c · W7
Ola 4 (gate)                 W9
```

### W0 — Toolchain (humano, 10 min) · bloquea todo
- Instalar Go 1.27.x (`scoop install go`), verificar `go version`.
- Instalar `wrangler` global pinneado; `surreal` CLI para local.
- **DoD:** `go build ./...` en `velo` y `openpassport` sale 0. Si falla, el error se convierte en el primer bug de W4.

### W1 — Verificador OpenPassport v1 · Ola 1
- **Archivos:** `pkg/manifest/**`, `pkg/manifest/testdata/**`.
- **Pasos:** reescribir `LoadAndVerify(dir)` según C1 (recibe directorio del pasaporte o ruta a `manifest.json`). Devolver struct con `name`, `packageVersion`, lista de archivos verificados. Copiar el template OP como fixture.
- **Tests:** golden OK; byte alterado en `01-IDENTITY.md` ⇒ error nombrando el archivo; self-hash alterado ⇒ error; CRLF en disco ⇒ OK (normaliza); entrada `manifest.json` ausente ⇒ error.
- **DoD:** `go test ./pkg/manifest -count=1` verde, 5 casos.

### W3 — Hotfix seguridad cockpit · Ola 1
- **Archivos:** `internal/scaffold/scaffold.go` (solo bloques `telegramAPI` y `envExample`).
- **Pasos:** aplicar C4 completo; quitar `INSFORGE_ANON_KEY` con valor de `.env.example` (dejar clave vacía); `compatibility_date` a fecha fija reciente.
- **DoD:** el TS generado no contiene `!==` sobre el secreto; no contiene ningún valor de `State` en archivos de secretos. Test Go que genera en `t.TempDir()` y hace `strings.Contains` de las garantías.
- *Nota:* W2a mueve este código a templates; W3 se mergea **antes** para que el fix viaje.

### W6 — CI · Ola 1
- **Archivos:** `.github/workflows/**`.
- **Pasos:** `ci.yml` en PR: `setup-go` con `go-version-file: go.mod`, `gofmt`, `vet`, `test`, build de la matriz de 4 targets. `release.yml` también usa `go-version-file`.
- **DoD:** workflow válido (`actionlint` limpio).

### W8 — OpenPassport: un solo contrato · Ola 1 · repo `openpassport`
- **Archivos:** `openpassport/pkg/**`, `openpassport/internal/passport/**`, tests.
- **Pasos:** eliminar/deprecar `pkg/schema/manifest.go` (Contract V0, H14). Exponer verificador público `pkg/verify` según C1. Si el builder Go usa `json.MarshalIndent`, usar `Encoder` con `SetEscapeHTML(false)` para igualar `JSON.stringify`. Test cruzado Go↔TS sobre el mismo fixture (TS: `node --test`).
- **DoD:** `go test ./...` y `pnpm -r test` verdes; mismo hash en ambos lenguajes.
- *Siguiente paso (fuera de ODD-001):* VELO importa `openpassport/pkg/verify` y borra W1. Hasta entonces, ambos comparten golden vector.

### W2a — Motor de templates · Ola 2
- **Archivos:** `internal/scaffold/**` (excepto `templates/data/*` y `templates/frontend/{next,react-vite}`).
- **Pasos:** migrar a C3 sin cambiar salida para `astro + insforge + telegram` (incluye fix de W3). Snapshot tests: render a `t.TempDir()` y comparar con `testdata/golden/<combo>/`. Flag de test `-update` para regenerar goldens.
- **DoD:** golden de la combinación actual byte-idéntica a la salida post-W3.

### W4 — CLI y wizard · Ola 2
- **Archivos:** `cmd/velo/**`, `internal/wizard/**`, `internal/tui/**`.
- **Pasos:** aplicar C2; flags `--frontend`, `--runtime`, `--data`, `--cockpit` validados contra enums (error + exit 2 si inválido); slugify en headless (H7); sin TTY y sin `--name` ⇒ error claro, exit 2 (H9); `--dry-run` lista archivos sin escribir; `velo doctor` (node, pnpm, wrangler, surreal: presente/versión); `velo verify <dir>` usa W1; paso BaaS → paso "Datos" con 3 opciones; borrar `Model.state` muerto (H13); exit codes: 0 ok / 1 runtime / 2 uso.
- **DoD:** tests de tabla para parseo de flags y slugify; `velo init --name "Mi App" --dry-run` imprime `mi-app/…`.

### W2b — Provider SurrealDB · Ola 3
- **Archivos:** `internal/scaffold/templates/data/surrealdb/**`, `data/none/**`.
- **Pasos:** `src/lib/db.ts` con SDK `surrealdb` JS (versión exacta pinneada; usar engine HTTP, compatible con Workers sin estado); `db/schema.surql` con tabla `pending_action` (id, kind, payload, status, expires_at) y `SCHEMAFULL`; `.dev.vars.example` con `SURREAL_URL`, `SURREAL_NS`, `SURREAL_DB`, `SURREAL_USER`, `SURREAL_PASS` vacíos; script `db:local` (`surreal start memory`), `db:migrate`. Cockpit (C4) resuelve approve/reject contra `pending_action`.
- **DoD:** golden test del combo `astro+surrealdb+telegram`; en W9 el proyecto compila.

### W2c — Providers de frontend · Ola 3
- **Archivos:** `internal/scaffold/templates/frontend/next/**`, `frontend/react-vite/**`.
- **Pasos:** Next.js vía `@opennextjs/cloudflare` (runtime Workers); React-Vite estático en Pages + `functions/api/telegram.ts`. Tailwind **3.4** fijado en ambos (regla del proyecto).
- **DoD:** golden tests por combo; cierra H6.

### W7 — Docs coherentes · Ola 3
- **Archivos:** `README.md`, `docs/**` (excepto este ODD).
- **Pasos:** cada afirmación del README debe estar respaldada por código o test (lista de claims ↔ archivo). SurrealDB como default, InsForge legacy. Badge Go = `go.mod`. Quitar "$0" donde dependa de un free tier no verificado; usar "free-tier first".
- **DoD:** tabla "Claim → Evidencia" al final del README sin huecos.

### W9 — Gate de integración · Ola 4
- **Archivos:** `scripts/e2e/**`, `.github/workflows/e2e.yml`.
- **Pasos:** para cada combo `{astro,next,react-vite} × {surrealdb,none}`: `velo init` headless en temp → `pnpm install --frozen-lockfile=false` → `pnpm build`. Smoke con `wrangler pages dev`/`wrangler dev`: `POST /api/telegram` sin secreto ⇒ 503, secreto malo ⇒ 401, owner incorrecto ⇒ 403. `velo verify` sobre template OP ⇒ 0; sobre copia alterada ⇒ ≠0.
- **Anti-Goodhart:** el gate falla si el PR modificó `**/testdata/**` o `*_test.go` sin `RATIONALE.md`.
- **DoD:** matriz verde → tag `v0.2.0`.

---

## 5. Fuera de alcance (backlog)
- Provider Python (Workers Python / FastAPI) y Go-en-edge (TinyGo/Wasm).
- MCP server de VELO (hoy el README lo implica; hoy no existe en este repo — vive en `openpassport/packages/mcp-passport`).
- `velo deploy` automatizado (wrangler + secrets).

## 6. Asignación de workers (costo mínimo)

Regla: el modelo más barato que pueda cumplir la DoD **sin decidir arquitectura**. Si el WP ya trae el contrato y el golden vector, es mecánico → Haiku.

| WP | Modelo | Esfuerzo | Por qué |
|---|---|---|---|
| W0 | Humano | — | Instalación local |
| W1 | Haiku 5.5 | low | Algoritmo + golden vector dados (C1) |
| W3 | Haiku 5.5 | low | Edición de strings acotada (C4) |
| W6 | Haiku 5.5 | low | YAML de CI estándar |
| W7 | Haiku 5.5 | low | Docs contra evidencia ya existente |
| W2a | Sonnet 5.5 | medium | Refactor con salida byte-idéntica |
| W4 | Sonnet 5.5 | medium | Muchas piezas pequeñas acopladas a C2 |
| W8 | Sonnet 5.5 | medium | Dos repos, Go↔TS |
| W2b · W2c | Sonnet 5.5 | medium | SDKs externos (SurrealDB, OpenNext) — riesgo de APIs inventadas |
| W9 | Sonnet 5.5 | medium | Scripts E2E + gate |
| Orquestador | Opus | — | Solo revisa `git diff --stat`, contratos y DoD entre olas; no escribe código |

Ahorro de tokens:
- Cada worker arranca **en contexto limpio** con solo: §2 (contratos), §3 (protocolo) y su WP. Nunca la conversación completa.
- Un worker que falla la DoD 2 veces **escala** (sube a Sonnet o vuelve al orquestador); no reintenta en bucle.
- Ola 1 (4 workers, 3 en Haiku) es la más barata y valida el método antes de gastar en Sonnet.

## 7. Coherencia con el discurso de entrevista
| Afirmación en el script | Hoy | Se cierra con |
|---|---|---|
| "fails closed" / HITL Approve antes del commit | Falso: fail-open, sin flujo approve | W3 + W2b |
| "Velo + OpenPassport, SHA-256" integrados | `verify` rechaza pasaportes reales | W1 + W8 |
| "No toy scripts / evalúo trayectoria, no exit 0" | 0 tests en VELO | W1–W4 tests + gate anti-Goodhart W9 |
| "commit en Postgres" | Migrando a SurrealDB | Decir "database transaction" (agnóstico) |
| "MVP en 48h" | Plausible tras W9 | W9 verde |
