# Guía de API

Base URL detrás del frontend: `/api`.

## Públicos

### `GET /api/health`

Comprueba API y conexión con PostgreSQL.

### `POST /api/auth/login`

```json
{
  "email": "admin@sena.edu.co",
  "password": "Admin123*"
}
```

Devuelve `access_token`, `refresh_token`, tiempo de expiración y usuario con roles/features efectivos.

### `POST /api/auth/refresh`

```json
{
  "refresh_token": "..."
}
```

Devuelve un access token nuevo si la sesión sigue vigente.

## Autenticados

Todos los siguientes usan:

```http
Authorization: Bearer ACCESS_TOKEN
```

### `GET /api/auth/me`

Perfil, roles y features del usuario autenticado.

### `POST /api/auth/logout`

Revoca el refresh token enviado.

### `GET /api/dashboard`

Requiere `DASH_CENTER_OVERVIEW`.

### `GET /api/users?q=texto&limit=100`

Requiere `IDENTITY_USER_VIEW`.

### `POST /api/users`

Requiere `IDENTITY_USER_MANAGE`. Si se envía `role_ids`, además se valida `IDENTITY_ROLE_ASSIGN`.

Ejemplo:

```json
{
  "email": "instructor.demo@sena.edu.co",
  "password": "Demo12345*",
  "first_name": "Laura",
  "last_name": "Gómez",
  "actor_type": "INSTRUCTOR",
  "role_ids": []
}
```

### `GET /api/users/{id}`

Requiere `IDENTITY_USER_VIEW`.

### `PUT /api/users/{id}`

Requiere `IDENTITY_USER_MANAGE`. Una contraseña vacía conserva la actual.

### `DELETE /api/users/{id}`

Requiere `IDENTITY_USER_MANAGE`. Desactiva al usuario y revoca sus sesiones.

### `PUT /api/users/{id}/roles`

Requiere `IDENTITY_ROLE_ASSIGN`.

```json
{
  "role_ids": ["uuid-del-rol"]
}
```

### `GET /api/roles`

Requiere `IDENTITY_ROLE_VIEW`.

### `GET /api/roles/{id}`

Detalle del rol con feature, módulo, acción y scope.

### `GET /api/catalog/modules`

Requiere `IDENTITY_ROLE_VIEW`. Devuelve módulos y features agrupadas.

### `GET /api/audit/logins?limit=200`

Requiere `AUDIT_LOG_VIEW`.

### `GET /api/sessions`

Devuelve las sesiones del usuario autenticado.
