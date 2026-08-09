# IAM DB

Identidad, autenticacion, RBAC (roles + features + scope) y sesiones.

> **Estado del pipeline: ✅ OK** — reestructurado, migraciones Liquibase corriendo limpio y verificado contra PostgreSQL 16 (2026-08-01).

## Estructura (Liquibase)

```
01_ddl/        DDL — 00_extensions, 01_schemas, 02_types, 03_tables (tablas sin FK),
               04_alter (llaves foraneas), 05_views..09_triggers, 10_indexes
02_dml/        Datos: 00_inserts (seeds/catalogos), 01_updates..04_patches
03_dcl/        Seguridad: 00_roles (reader/writer/admin), 01_grants, 02_policies
04_tcl/        Tags de version / release
05_rollbacks/  Rollbacks espejo de cada changeset
changelog/     changelog-master.yaml (punto de entrada del pipeline)
```

- **Schema(s) propio(s):** `identity, rbac, rbac_catalog, session, identity_audit` — ninguna tabla queda en `public` (aislamiento por microservicio).
- **Regla clave:** las tablas se crean en `03_tables` **sin FKs**; las llaves foraneas se agregan en `04_alter` (evita rupturas por orden de creacion).

## Como correr

Desde `design-software/docker-infra` (orquestador unico, base de datos compartida):

```bash
docker compose --env-file .env.develop up postgres -d
docker compose --env-file .env.develop --profile tooling run --rm liquibase-iam update
```

Estado / rollback:

```bash
docker compose --profile tooling run --rm liquibase-iam status --verbose
docker compose --profile tooling run --rm liquibase-iam rollbackCount 1
```

## Verificacion realizada

- `update` aplica **desde cero** sin errores en los 4 ambientes (develop / qa / staging / main).
- Comportamiento **incremental** confirmado: un changeset nuevo corre solo el nuevo (`Previously run` intacto).
- **Rollback** verificado.
- Baseline del trabajo previo congelado en el tag **`1.0.0`**.
