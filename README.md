# IAM Console — Backend Go + Frontend React

Aplicación funcional construida sobre la base de datos **IAM DB** incluida en este repositorio. El producto usa la estructura PostgreSQL original y no reemplaza sus migraciones Liquibase.

## Qué incluye

- **Backend REST en Go** (`net/http` + `database/sql` + PostgreSQL).
- **Frontend en React** con Vite y diseño responsive.
- **PostgreSQL 16**.
- **Liquibase** ejecutando el changelog original del ZIP.
- **Docker Compose** para levantar todo con un comando.
- Autenticación con access token JWT + refresh token persistido.
- Contraseñas con PBKDF2-HMAC-SHA256 y salt aleatorio.
- Bloqueo temporal después de 5 intentos fallidos.
- Auditoría de inicios de sesión exitosos y fallidos.
- CRUD operativo de usuarios (la eliminación es desactivación segura).
- Asignación de roles RBAC.
- Consulta de roles, módulos, features, scopes y niveles de acción.
- Panel con métricas de identidad y sesiones.
- Consulta de las sesiones propias.

## Qué representa realmente la base de datos

La base entregada corresponde al dominio **IAM (Identity and Access Management)**. Contiene 5 esquemas:

| Esquema | Responsabilidad |
|---|---|
| `identity` | Usuarios e identidad básica |
| `rbac_catalog` | Módulos y features/permisos disponibles |
| `rbac` | Roles, asignaciones, scopes y overrides |
| `session` | Refresh tokens y solicitudes de recuperación |
| `identity_audit` | Auditoría de autenticación |

El seed contiene **10 módulos**, **7 roles** y **62 features**. Algunos módulos se llaman Gestión Académica, Horarios, Documentos o Monitoreo, pero en esta BD son **catálogos de autorización** para otros servicios. El ZIP no contiene las tablas de fichas, horarios, documentos, etc.; por eso este producto administra identidades y permisos sin inventar datos de otros microservicios.

## Inicio rápido con Docker

Requisitos:

- Docker Desktop / Docker Engine con Docker Compose.
- Puertos `3000` y `5432` disponibles (se pueden cambiar en `.env`).

### 1. Configurar variables

En PowerShell:

```powershell
Copy-Item .env.example .env
```

En Linux/macOS:

```bash
cp .env.example .env
```

Para una demostración local puedes dejar los valores por defecto. Para un entorno real cambia al menos `POSTGRES_PASSWORD`, `JWT_SECRET` y `BOOTSTRAP_ADMIN_PASSWORD`.

### 2. Levantar todo

```bash
docker compose up --build
```

Docker hará lo siguiente, en orden:

1. Levanta PostgreSQL 16.
2. Espera a que PostgreSQL esté saludable.
3. Ejecuta el `changelog-master.yaml` original con Liquibase.
4. Compila y levanta la API de Go.
5. Inicializa/repara el administrador demo y garantiza que `SYSTEM_ADMIN` tenga acceso al catálogo IAM completo.
6. Compila React y lo sirve con Nginx.

### 3. Abrir el producto

Navegador:

```text
http://localhost:3000
```

Credenciales demo por defecto:

```text
Correo:     admin@sena.edu.co
Contraseña: Admin123*
```

> El seed original trae un `password_hash` demostrativo que no es una contraseña válida. El backend detecta ese hash inicial y lo sustituye por un hash real PBKDF2 en el primer arranque. Después no pisa una contraseña ya válida en cada reinicio.

## Detener y reiniciar

Detener contenedores:

```bash
docker compose down
```

Volver a iniciar:

```bash
docker compose up
```

Reconstruir después de cambios de código:

```bash
docker compose up --build
```

Borrar también la base de datos y empezar desde cero:

```bash
docker compose down -v
docker compose up --build
```

## Desarrollo sin empaquetar el frontend

Primero puedes dejar PostgreSQL y Liquibase en Docker:

```bash
docker compose up postgres liquibase
```

Backend:

```bash
cd backend
go mod download
go run ./cmd/api
```

Frontend, en otra terminal:

```bash
cd frontend
npm install
npm run dev
```

Vite abre el frontend en `http://localhost:5173` y redirige `/api` al backend en `http://localhost:8080`.

## Pantallas del frontend

### Inicio de sesión

Autentica contra `identity.user`, registra el intento en `identity_audit.audit_login`, actualiza `last_login_at` y genera tokens de sesión.

### Panel principal

Muestra:

- usuarios registrados;
- usuarios activos;
- roles disponibles;
- refresh tokens vigentes;
- intentos fallidos en las últimas 24 horas.

### Usuarios

Permite buscar, crear, editar y desactivar usuarios. También permite asignar roles si el usuario autenticado tiene `IDENTITY_ROLE_ASSIGN`.

### Roles

Muestra los 7 roles del seed y, al abrir uno, lista cada feature, su módulo, `action_level` y `scope_type`.

### Permisos

Visualiza los 10 módulos y sus 62 features directamente desde `rbac_catalog`.

### Auditoría

Muestra los eventos de `identity_audit.audit_login`, incluyendo resultado, IP y user-agent.

### Mis sesiones

Muestra los refresh tokens del usuario autenticado y si cada sesión continúa activa, fue revocada o expiró.

## Seguridad y RBAC

El token JWT identifica al usuario, pero **los permisos efectivos no se confían ciegamente al JWT**. En cada petición autenticada el backend vuelve a consultar roles/features vigentes en PostgreSQL. Así, si un administrador cambia un rol, el cambio puede aplicarse sin esperar a que caduque el access token.

Ejemplos de autorización:

| Endpoint | Permiso requerido |
|---|---|
| `GET /api/users` | `IDENTITY_USER_VIEW` |
| `POST /api/users` | `IDENTITY_USER_MANAGE` |
| `PUT /api/users/{id}/roles` | `IDENTITY_ROLE_ASSIGN` |
| `GET /api/roles` | `IDENTITY_ROLE_VIEW` |
| `GET /api/catalog/modules` | `IDENTITY_ROLE_VIEW` |
| `GET /api/audit/logins` | `AUDIT_LOG_VIEW` |
| `GET /api/dashboard` | `DASH_CENTER_OVERVIEW` |

El backend también evita que un administrador se quite sus propios roles desde la API, reduciendo el riesgo de autobloqueo accidental.

## Estructura

```text
IAM-Sena/
├── backend/
│   ├── cmd/api/main.go
│   ├── internal/app/
│   │   ├── app.go
│   │   └── security.go
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── App.jsx
│   │   ├── api.js
│   │   ├── main.jsx
│   │   └── styles.css
│   ├── nginx/default.conf
│   ├── Dockerfile
│   └── package.json
├── database/               # Base de datos original recibida
├── docs/ 
│   └── GUIA_API.md
├── docker-compose.yml
└── .env.example
```

## Comprobaciones rápidas

Salud de la API a través del frontend/Nginx:

```bash
curl http://localhost:3000/api/health
```

Respuesta esperada:

```json
{"database":true,"service":"iam-api","status":"ok"}
```

Ver contenedores:

```bash
docker compose ps
```

Ver logs de la API:

```bash
docker compose logs -f backend
```

Ver migraciones Liquibase:

```bash
docker compose logs liquibase
```

## Validación realizada durante la generación

- Se inspeccionaron las tablas, FKs, índices, schemas y seeds del ZIP recibido.
- Se verificó que el backend Go compila a nivel de código.
- Se verificó sintácticamente el JSX/JavaScript del frontend.
- El entorno de generación no dispone de Docker/PostgreSQL ejecutables, por lo que la prueba de integración con contenedores debe ejecutarse en una máquina con Docker usando los comandos anteriores.
