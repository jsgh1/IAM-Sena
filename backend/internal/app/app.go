package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	AdminEmail     string
	AdminPassword  string
	AdminFirstName string
	AdminLastName  string
}

func LoadConfig() Config {
	return Config{
		Port:           env("PORT", "8080"),
		DatabaseURL:    env("DATABASE_URL", "postgres://iam:iam@localhost:5432/iam?sslmode=disable"),
		JWTSecret:      env("JWT_SECRET", "change-this-development-secret"),
		AccessTTL:      durationEnv("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTTL:     durationEnv("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		AdminEmail:     env("BOOTSTRAP_ADMIN_EMAIL", "admin@sena.edu.co"),
		AdminPassword:  env("BOOTSTRAP_ADMIN_PASSWORD", "Admin123*"),
		AdminFirstName: env("BOOTSTRAP_ADMIN_FIRST_NAME", "System"),
		AdminLastName:  env("BOOTSTRAP_ADMIN_LAST_NAME", "Admin"),
	}
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func durationEnv(k string, d time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if x, e := time.ParseDuration(v); e == nil {
			return x
		}
	}
	return d
}

type App struct {
	db  *sql.DB
	cfg Config
}

type apiError struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

type authUser struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	ActorType string   `json:"actor_type"`
	IsActive  bool     `json:"is_active"`
	Roles     []string `json:"roles"`
	Features  []string `json:"features"`
}

type contextKey string

const userContextKey contextKey = "auth-user"

func New(cfg Config) (*App, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	for {
		if err = db.PingContext(ctx); err == nil {
			break
		}
		if ctx.Err() != nil {
			db.Close()
			return nil, fmt.Errorf("database no disponible: %w", err)
		}
		time.Sleep(750 * time.Millisecond)
	}
	a := &App{db: db, cfg: cfg}
	if err := a.bootstrapAdmin(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("bootstrap admin: %w", err)
	}
	return a, nil
}

func (a *App) Close() error { return a.db.Close() }

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", a.health)
	mux.HandleFunc("POST /api/auth/login", a.login)
	mux.HandleFunc("POST /api/auth/refresh", a.refresh)
	mux.Handle("POST /api/auth/logout", a.auth(http.HandlerFunc(a.logout)))
	mux.Handle("GET /api/auth/me", a.auth(http.HandlerFunc(a.me)))
	mux.Handle("GET /api/dashboard", a.auth(a.require("DASH_CENTER_OVERVIEW", http.HandlerFunc(a.dashboard))))
	mux.Handle("GET /api/users", a.auth(a.require("IDENTITY_USER_VIEW", http.HandlerFunc(a.listUsers))))
	mux.Handle("POST /api/users", a.auth(a.require("IDENTITY_USER_MANAGE", http.HandlerFunc(a.createUser))))
	mux.Handle("GET /api/users/{id}", a.auth(a.require("IDENTITY_USER_VIEW", http.HandlerFunc(a.getUser))))
	mux.Handle("PUT /api/users/{id}", a.auth(a.require("IDENTITY_USER_MANAGE", http.HandlerFunc(a.updateUser))))
	mux.Handle("DELETE /api/users/{id}", a.auth(a.require("IDENTITY_USER_MANAGE", http.HandlerFunc(a.deactivateUser))))
	mux.Handle("PUT /api/users/{id}/roles", a.auth(a.require("IDENTITY_ROLE_ASSIGN", http.HandlerFunc(a.setUserRoles))))
	mux.Handle("GET /api/roles", a.auth(a.require("IDENTITY_ROLE_VIEW", http.HandlerFunc(a.listRoles))))
	mux.Handle("GET /api/roles/{id}", a.auth(a.require("IDENTITY_ROLE_VIEW", http.HandlerFunc(a.getRole))))
	mux.Handle("GET /api/catalog/modules", a.auth(a.require("IDENTITY_ROLE_VIEW", http.HandlerFunc(a.listModules))))
	mux.Handle("GET /api/audit/logins", a.auth(a.require("AUDIT_LOG_VIEW", http.HandlerFunc(a.auditLogins))))
	mux.Handle("GET /api/sessions", a.auth(http.HandlerFunc(a.sessions)))
	return a.middleware(mux)
}

func (a *App) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		defer func() {
			if x := recover(); x != nil {
				log.Printf("panic: %v", x)
				writeJSON(w, 500, apiError{Error: "error interno"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *App) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			writeJSON(w, 401, apiError{Error: "autenticacion requerida"})
			return
		}
		claims, err := parseAccessToken(a.cfg.JWTSecret, strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			writeJSON(w, 401, apiError{Error: "sesion expirada o token invalido"})
			return
		}
		u, err := a.loadAuthUser(r.Context(), claims.Sub)
		if err != nil || !u.IsActive {
			writeJSON(w, 401, apiError{Error: "usuario no disponible"})
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, u)))
	})
}

func (a *App) require(feature string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := currentUser(r)
		for _, f := range u.Features {
			if f == feature {
				next.ServeHTTP(w, r)
				return
			}
		}
		writeJSON(w, 403, apiError{Error: "permiso insuficiente", Details: feature})
	})
}

func currentUser(r *http.Request) authUser {
	u, _ := r.Context().Value(userContextKey).(authUser)
	return u
}
func hasFeature(u authUser, code string) bool {
	for _, f := range u.Features {
		if f == code {
			return true
		}
	}
	return false
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeJSON(w, 400, apiError{Error: "JSON invalido", Details: err.Error()})
		return false
	}
	return true
}
func cleanEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
func clientIP(r *http.Request) string {
	if x := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); x != "" {
		return x
	}
	if i := strings.LastIndex(r.RemoteAddr, ":"); i > 0 {
		return strings.Trim(r.RemoteAddr[:i], "[]")
	}
	return r.RemoteAddr
}

func (a *App) bootstrapAdmin(ctx context.Context) error {
	hash, err := hashPassword(a.cfg.AdminPassword)
	if err != nil {
		return err
	}
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id, currentHash string
	err = tx.QueryRowContext(ctx, `SELECT id,password_hash FROM identity."user" WHERE lower(email)=lower($1)`, a.cfg.AdminEmail).Scan(&id, &currentHash)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `INSERT INTO identity."user" (email,password_hash,first_name,last_name,actor_type,is_active)
			VALUES ($1,$2,$3,$4,'USER',true) RETURNING id`, cleanEmail(a.cfg.AdminEmail), hash, a.cfg.AdminFirstName, a.cfg.AdminLastName).Scan(&id)
	} else if err == nil {
		if !strings.HasPrefix(currentHash, "pbkdf2_sha256$") {
			_, err = tx.ExecContext(ctx, `UPDATE identity."user" SET email=$2,password_hash=$3,first_name=$4,last_name=$5,is_active=true,updated_at=now() WHERE id=$1`, id, cleanEmail(a.cfg.AdminEmail), hash, a.cfg.AdminFirstName, a.cfg.AdminLastName)
		} else {
			_, err = tx.ExecContext(ctx, `UPDATE identity."user" SET first_name=$2,last_name=$3,is_active=true,updated_at=now() WHERE id=$1`, id, a.cfg.AdminFirstName, a.cfg.AdminLastName)
		}
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO rbac.user_role (user_id,role_id,assigned_by)
        SELECT $1,r.id,$1 FROM rbac.role r WHERE r.name='SYSTEM_ADMIN'
        ON CONFLICT DO NOTHING`, id)
	if err != nil {
		return err
	}
	// El administrador del sistema debe poder operar todo el catalogo IAM.
	// El seed original define algunas features sin asociarlas a ningun rol.
	_, err = tx.ExecContext(ctx, `INSERT INTO rbac.role_feature(role_id,feature_id,scope_type)
		SELECT r.id,f.id,'GLOBAL' FROM rbac.role r CROSS JOIN rbac_catalog.feature f
		WHERE r.name='SYSTEM_ADMIN' AND f.is_active=true
		ON CONFLICT (role_id,feature_id) DO NOTHING`)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (a *App) loadAuthUser(ctx context.Context, id string) (authUser, error) {
	var u authUser
	err := a.db.QueryRowContext(ctx, `SELECT id,email,first_name,last_name,actor_type,is_active FROM identity."user" WHERE id=$1`, id).Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.ActorType, &u.IsActive)
	if err != nil {
		return u, err
	}
	rows, err := a.db.QueryContext(ctx, `SELECT DISTINCT r.name FROM rbac.user_role ur JOIN rbac.role r ON r.id=ur.role_id WHERE ur.user_id=$1 AND (ur.expires_at IS NULL OR ur.expires_at>now()) ORDER BY r.name`, id)
	if err != nil {
		return u, err
	}
	defer rows.Close()
	for rows.Next() {
		var x string
		if err := rows.Scan(&x); err != nil {
			return u, err
		}
		u.Roles = append(u.Roles, x)
	}
	rows2, err := a.db.QueryContext(ctx, `SELECT DISTINCT f.code FROM rbac.user_role ur JOIN rbac.role_feature rf ON rf.role_id=ur.role_id JOIN rbac_catalog.feature f ON f.id=rf.feature_id WHERE ur.user_id=$1 AND f.is_active=true AND (ur.expires_at IS NULL OR ur.expires_at>now())
        UNION SELECT DISTINCT f.code FROM rbac.user_scope_override o JOIN rbac_catalog.feature f ON f.id=o.feature_id WHERE o.user_id=$1 AND o.is_allowed=true AND (o.expires_at IS NULL OR o.expires_at>now()) ORDER BY 1`, id)
	if err != nil {
		return u, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var x string
		if err := rows2.Scan(&x); err != nil {
			return u, err
		}
		u.Features = append(u.Features, x)
	}
	return u, rows2.Err()
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := a.db.PingContext(ctx); err != nil {
		writeJSON(w, 503, map[string]any{"status": "error", "database": false})
		return
	}
	writeJSON(w, 200, map[string]any{"status": "ok", "database": true, "service": "iam-api"})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var in loginRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	in.Email = cleanEmail(in.Email)
	var id, email, hash, first, last, actor string
	var active bool
	var failed int
	var locked sql.NullTime
	err := a.db.QueryRowContext(r.Context(), `SELECT id,email,password_hash,first_name,last_name,actor_type,is_active,failed_attempts,locked_until FROM identity."user" WHERE lower(email)=lower($1)`, in.Email).Scan(&id, &email, &hash, &first, &last, &actor, &active, &failed, &locked)
	if errors.Is(err, sql.ErrNoRows) {
		a.audit(r.Context(), nil, in.Email, "USER_NOT_FOUND", r)
		writeJSON(w, 401, apiError{Error: "credenciales invalidas"})
		return
	}
	if err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudo iniciar sesion"})
		return
	}
	if locked.Valid && locked.Time.After(time.Now()) {
		a.audit(r.Context(), &id, email, "ACCOUNT_LOCKED", r)
		writeJSON(w, 423, apiError{Error: "cuenta bloqueada temporalmente"})
		return
	}
	if !active {
		a.audit(r.Context(), &id, email, "ACCOUNT_LOCKED", r)
		writeJSON(w, 403, apiError{Error: "cuenta inactiva"})
		return
	}
	if !verifyPassword(hash, in.Password) {
		failed++
		var until any = nil
		if failed >= 5 {
			until = time.Now().Add(15 * time.Minute)
			failed = 0
		}
		_, _ = a.db.ExecContext(r.Context(), `UPDATE identity."user" SET failed_attempts=$2,locked_until=$3,updated_at=now() WHERE id=$1`, id, failed, until)
		a.audit(r.Context(), &id, email, "INVALID_PASSWORD", r)
		writeJSON(w, 401, apiError{Error: "credenciales invalidas"})
		return
	}
	_, _ = a.db.ExecContext(r.Context(), `UPDATE identity."user" SET failed_attempts=0,locked_until=NULL,last_login_at=now(),updated_at=now() WHERE id=$1`, id)
	a.audit(r.Context(), &id, email, "SUCCESS", r)
	access, _ := signAccessToken(a.cfg.JWTSecret, id, email, a.cfg.AccessTTL)
	refresh, refreshHash, err := randomToken()
	if err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudo crear sesion"})
		return
	}
	_, err = a.db.ExecContext(r.Context(), `INSERT INTO session.refresh_token(user_id,token_hash,device_hint,ip_address,expires_at) VALUES($1,$2,$3,$4,$5)`, id, refreshHash, trimLen(r.UserAgent(), 200), clientIP(r), time.Now().Add(a.cfg.RefreshTTL))
	if err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudo guardar sesion"})
		return
	}
	u, _ := a.loadAuthUser(r.Context(), id)
	writeJSON(w, 200, map[string]any{"access_token": access, "refresh_token": refresh, "expires_in": int(a.cfg.AccessTTL.Seconds()), "user": u})
}

func trimLen(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}
func (a *App) audit(ctx context.Context, userID *string, email, outcome string, r *http.Request) {
	var uid any = nil
	if userID != nil {
		uid = *userID
	}
	_, _ = a.db.ExecContext(ctx, `INSERT INTO identity_audit.audit_login(user_id,email_attempted,outcome,ip_address,user_agent) VALUES($1,$2,$3,$4,$5)`, uid, email, outcome, clientIP(r), r.UserAgent())
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (a *App) refresh(w http.ResponseWriter, r *http.Request) {
	var in refreshRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	h := tokenHash(in.RefreshToken)
	var id, email string
	err := a.db.QueryRowContext(r.Context(), `SELECT u.id,u.email FROM session.refresh_token t JOIN identity."user" u ON u.id=t.user_id WHERE t.token_hash=$1 AND t.is_revoked=false AND t.expires_at>now() AND u.is_active=true`, h).Scan(&id, &email)
	if err != nil {
		writeJSON(w, 401, apiError{Error: "refresh token invalido o expirado"})
		return
	}
	access, _ := signAccessToken(a.cfg.JWTSecret, id, email, a.cfg.AccessTTL)
	writeJSON(w, 200, map[string]any{"access_token": access, "expires_in": int(a.cfg.AccessTTL.Seconds())})
}
func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	var in refreshRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	_, _ = a.db.ExecContext(r.Context(), `UPDATE session.refresh_token SET is_revoked=true,revoked_at=now() WHERE token_hash=$1`, tokenHash(in.RefreshToken))
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) me(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, currentUser(r)) }

func (a *App) dashboard(w http.ResponseWriter, r *http.Request) {
	var users, active, roles, sessions, failed int
	_ = a.db.QueryRowContext(r.Context(), `SELECT count(*) FROM identity."user"`).Scan(&users)
	_ = a.db.QueryRowContext(r.Context(), `SELECT count(*) FROM identity."user" WHERE is_active=true`).Scan(&active)
	_ = a.db.QueryRowContext(r.Context(), `SELECT count(*) FROM rbac.role`).Scan(&roles)
	_ = a.db.QueryRowContext(r.Context(), `SELECT count(*) FROM session.refresh_token WHERE is_revoked=false AND expires_at>now()`).Scan(&sessions)
	_ = a.db.QueryRowContext(r.Context(), `SELECT count(*) FROM identity_audit.audit_login WHERE outcome<>'SUCCESS' AND attempted_at>now()-interval '24 hours'`).Scan(&failed)
	writeJSON(w, 200, map[string]any{"users": users, "active_users": active, "roles": roles, "active_sessions": sessions, "failed_logins_24h": failed})
}

func (a *App) listUsers(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := parseLimit(r.URL.Query().Get("limit"), 100)
	rows, err := a.db.QueryContext(r.Context(), `SELECT u.id,u.email,u.first_name,u.last_name,u.actor_type,u.is_active,u.last_login_at,u.created_at,COALESCE(string_agg(DISTINCT ro.display_name,', ' ORDER BY ro.display_name),'') FROM identity."user" u LEFT JOIN rbac.user_role ur ON ur.user_id=u.id LEFT JOIN rbac.role ro ON ro.id=ur.role_id WHERE ($1='' OR u.email ILIKE '%'||$1||'%' OR u.first_name ILIKE '%'||$1||'%' OR u.last_name ILIKE '%'||$1||'%') GROUP BY u.id ORDER BY u.created_at DESC LIMIT $2`, q, limit)
	if err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudieron listar usuarios"})
		return
	}
	defer rows.Close()
	type row struct {
		ID, Email, FirstName, LastName, ActorType string
		IsActive                                  bool
		LastLoginAt                               *time.Time
		CreatedAt                                 time.Time
		Roles                                     string
	}
	out := []row{}
	for rows.Next() {
		var x row
		var ll sql.NullTime
		if err := rows.Scan(&x.ID, &x.Email, &x.FirstName, &x.LastName, &x.ActorType, &x.IsActive, &ll, &x.CreatedAt, &x.Roles); err != nil {
			writeJSON(w, 500, apiError{Error: "error leyendo usuarios"})
			return
		}
		if ll.Valid {
			x.LastLoginAt = &ll.Time
		}
		out = append(out, x)
	}
	writeJSON(w, 200, out)
}
func parseLimit(s string, d int) int {
	n, e := strconv.Atoi(s)
	if e != nil || n < 1 {
		return d
	}
	if n > 500 {
		return 500
	}
	return n
}

func sameStringSet(a, b []string) bool {
	ma := map[string]struct{}{}
	mb := map[string]struct{}{}
	for _, x := range a {
		ma[x] = struct{}{}
	}
	for _, x := range b {
		mb[x] = struct{}{}
	}
	if len(ma) != len(mb) {
		return false
	}
	for x := range ma {
		if _, ok := mb[x]; !ok {
			return false
		}
	}
	return true
}

type userInput struct {
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	ActorType string   `json:"actor_type"`
	IsActive  *bool    `json:"is_active,omitempty"`
	RoleIDs   []string `json:"role_ids,omitempty"`
}

func validActor(s string) bool { return s == "USER" || s == "INSTRUCTOR" || s == "LEARNER" }
func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var in userInput
	if !decodeJSON(w, r, &in) {
		return
	}
	in.Email = cleanEmail(in.Email)
	if in.Email == "" || len(in.Password) < 8 || strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" || !validActor(in.ActorType) {
		writeJSON(w, 422, apiError{Error: "datos incompletos; password minimo 8 caracteres"})
		return
	}
	if len(in.RoleIDs) > 0 && !hasFeature(currentUser(r), "IDENTITY_ROLE_ASSIGN") {
		writeJSON(w, 403, apiError{Error: "no tienes permiso para asignar roles"})
		return
	}
	hash, _ := hashPassword(in.Password)
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudo crear usuario"})
		return
	}
	defer tx.Rollback()
	var id string
	err = tx.QueryRowContext(r.Context(), `INSERT INTO identity."user"(email,password_hash,first_name,last_name,actor_type) VALUES($1,$2,$3,$4,$5) RETURNING id`, in.Email, hash, strings.TrimSpace(in.FirstName), strings.TrimSpace(in.LastName), in.ActorType).Scan(&id)
	if err != nil {
		writeJSON(w, 409, apiError{Error: "el correo ya existe o los datos no son validos", Details: err.Error()})
		return
	}
	admin := currentUser(r)
	for _, rid := range in.RoleIDs {
		if _, err = tx.ExecContext(r.Context(), `INSERT INTO rbac.user_role(user_id,role_id,assigned_by) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, id, rid, admin.ID); err != nil {
			writeJSON(w, 422, apiError{Error: "rol invalido"})
			return
		}
	}
	if err = tx.Commit(); err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudo confirmar usuario"})
		return
	}
	writeJSON(w, 201, map[string]string{"id": id})
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	u, err := a.loadAuthUser(r.Context(), id)
	if err != nil {
		writeJSON(w, 404, apiError{Error: "usuario no encontrado"})
		return
	}
	writeJSON(w, 200, u)
}
func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in userInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if !validActor(in.ActorType) {
		writeJSON(w, 422, apiError{Error: "actor_type invalido"})
		return
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	var res sql.Result
	var err error
	if strings.TrimSpace(in.Password) != "" {
		if len(in.Password) < 8 {
			writeJSON(w, 422, apiError{Error: "password minimo 8 caracteres"})
			return
		}
		hash, _ := hashPassword(in.Password)
		res, err = a.db.ExecContext(r.Context(), `UPDATE identity."user" SET email=$2,first_name=$3,last_name=$4,actor_type=$5,is_active=$6,password_hash=$7,updated_at=now() WHERE id=$1`, id, cleanEmail(in.Email), strings.TrimSpace(in.FirstName), strings.TrimSpace(in.LastName), in.ActorType, active, hash)
	} else {
		res, err = a.db.ExecContext(r.Context(), `UPDATE identity."user" SET email=$2,first_name=$3,last_name=$4,actor_type=$5,is_active=$6,updated_at=now() WHERE id=$1`, id, cleanEmail(in.Email), strings.TrimSpace(in.FirstName), strings.TrimSpace(in.LastName), in.ActorType, active)
	}
	if err != nil {
		writeJSON(w, 409, apiError{Error: "no se pudo actualizar usuario", Details: err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, 404, apiError{Error: "usuario no encontrado"})
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) deactivateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == currentUser(r).ID {
		writeJSON(w, 422, apiError{Error: "no puedes desactivar tu propia cuenta"})
		return
	}
	res, err := a.db.ExecContext(r.Context(), `UPDATE identity."user" SET is_active=false,updated_at=now() WHERE id=$1`, id)
	if err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudo desactivar"})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, 404, apiError{Error: "usuario no encontrado"})
		return
	}
	_, _ = a.db.ExecContext(r.Context(), `UPDATE session.refresh_token SET is_revoked=true,revoked_at=now() WHERE user_id=$1 AND is_revoked=false`, id)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

type roleSet struct {
	RoleIDs []string `json:"role_ids"`
}

func (a *App) setUserRoles(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in roleSet
	if !decodeJSON(w, r, &in) {
		return
	}
	admin := currentUser(r)
	if id == admin.ID {
		rows, err := a.db.QueryContext(r.Context(), `SELECT role_id FROM rbac.user_role WHERE user_id=$1 AND (expires_at IS NULL OR expires_at>now())`, id)
		if err != nil {
			writeJSON(w, 500, apiError{Error: "no se pudieron validar tus roles"})
			return
		}
		current := []string{}
		for rows.Next() {
			var rid string
			if rows.Scan(&rid) == nil {
				current = append(current, rid)
			}
		}
		rows.Close()
		if !sameStringSet(current, in.RoleIDs) {
			writeJSON(w, 422, apiError{Error: "no puedes cambiar tus propios roles"})
			return
		}
		writeJSON(w, 200, map[string]bool{"ok": true})
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudieron actualizar roles"})
		return
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(r.Context(), `DELETE FROM rbac.user_role WHERE user_id=$1`, id); err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudieron limpiar roles"})
		return
	}
	for _, rid := range in.RoleIDs {
		if _, err = tx.ExecContext(r.Context(), `INSERT INTO rbac.user_role(user_id,role_id,assigned_by) VALUES($1,$2,$3)`, id, rid, admin.ID); err != nil {
			writeJSON(w, 422, apiError{Error: "rol invalido", Details: rid})
			return
		}
	}
	if err = tx.Commit(); err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudieron guardar roles"})
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) listRoles(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), `SELECT r.id,r.name,r.display_name,r.description,r.is_system_role,count(rf.id) FROM rbac.role r LEFT JOIN rbac.role_feature rf ON rf.role_id=r.id GROUP BY r.id ORDER BY r.display_name`)
	if err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudieron listar roles"})
		return
	}
	defer rows.Close()
	type rr struct {
		ID, Name, DisplayName string
		Description           *string
		IsSystemRole          bool
		FeatureCount          int
	}
	out := []rr{}
	for rows.Next() {
		var x rr
		var d sql.NullString
		if err := rows.Scan(&x.ID, &x.Name, &x.DisplayName, &d, &x.IsSystemRole, &x.FeatureCount); err != nil {
			return
		}
		if d.Valid {
			x.Description = &d.String
		}
		out = append(out, x)
	}
	writeJSON(w, 200, out)
}
func (a *App) getRole(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var out struct {
		ID          string           `json:"id"`
		Name        string           `json:"name"`
		DisplayName string           `json:"display_name"`
		Description *string          `json:"description"`
		Features    []map[string]any `json:"features"`
	}
	var d sql.NullString
	err := a.db.QueryRowContext(r.Context(), `SELECT id,name,display_name,description FROM rbac.role WHERE id=$1`, id).Scan(&out.ID, &out.Name, &out.DisplayName, &d)
	if err != nil {
		writeJSON(w, 404, apiError{Error: "rol no encontrado"})
		return
	}
	if d.Valid {
		out.Description = &d.String
	}
	rows, err := a.db.QueryContext(r.Context(), `SELECT f.code,f.name,m.name,rf.scope_type,f.action_level FROM rbac.role_feature rf JOIN rbac_catalog.feature f ON f.id=rf.feature_id JOIN rbac_catalog.module m ON m.id=f.module_id WHERE rf.role_id=$1 ORDER BY m.display_order,f.name`, id)
	if err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudieron cargar permisos"})
		return
	}
	defer rows.Close()
	out.Features = []map[string]any{}
	for rows.Next() {
		var code, name, module, scope, action string
		_ = rows.Scan(&code, &name, &module, &scope, &action)
		out.Features = append(out.Features, map[string]any{"code": code, "name": name, "module": module, "scope": scope, "action": action})
	}
	writeJSON(w, 200, out)
}
func (a *App) listModules(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), `SELECT m.id,m.code,m.name,m.description,m.display_order,m.icon_key,m.is_active,f.id,f.code,f.name,f.description,f.action_level,f.is_active FROM rbac_catalog.module m LEFT JOIN rbac_catalog.feature f ON f.module_id=m.id ORDER BY m.display_order,f.name`)
	if err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudo cargar catalogo"})
		return
	}
	defer rows.Close()
	type feature struct {
		ID, Code, Name, Action string
		Description            *string
		Active                 bool
	}
	type module struct {
		ID, Code, Name string
		Description    *string
		Order          int
		Icon           *string
		Active         bool
		Features       []feature
	}
	mods := []module{}
	idx := map[string]int{}
	for rows.Next() {
		var mid, mcode, mname string
		var mdesc, micon sql.NullString
		var order int
		var mactive bool
		var fid, fcode, fname, faction sql.NullString
		var fdesc sql.NullString
		var factive sql.NullBool
		if err := rows.Scan(&mid, &mcode, &mname, &mdesc, &order, &micon, &mactive, &fid, &fcode, &fname, &fdesc, &faction, &factive); err != nil {
			writeJSON(w, 500, apiError{Error: "error leyendo catalogo"})
			return
		}
		i, ok := idx[mid]
		if !ok {
			m := module{ID: mid, Code: mcode, Name: mname, Order: order, Active: mactive, Features: []feature{}}
			if mdesc.Valid {
				m.Description = &mdesc.String
			}
			if micon.Valid {
				m.Icon = &micon.String
			}
			mods = append(mods, m)
			i = len(mods) - 1
			idx[mid] = i
		}
		if fid.Valid {
			f := feature{ID: fid.String, Code: fcode.String, Name: fname.String, Action: faction.String, Active: factive.Bool}
			if fdesc.Valid {
				f.Description = &fdesc.String
			}
			mods[i].Features = append(mods[i].Features, f)
		}
	}
	writeJSON(w, 200, mods)
}

func (a *App) auditLogins(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r.URL.Query().Get("limit"), 100)
	rows, err := a.db.QueryContext(r.Context(), `SELECT id,user_id,email_attempted,outcome,ip_address,user_agent,attempted_at FROM identity_audit.audit_login ORDER BY attempted_at DESC LIMIT $1`, limit)
	if err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudo consultar auditoria"})
		return
	}
	defer rows.Close()
	type row struct {
		ID                   string  `json:"id"`
		UserID               *string `json:"user_id"`
		Email, Outcome       string
		IPAddress, UserAgent *string
		AttemptedAt          time.Time
	}
	out := []row{}
	for rows.Next() {
		var x row
		var uid, ip, ua sql.NullString
		if err := rows.Scan(&x.ID, &uid, &x.Email, &x.Outcome, &ip, &ua, &x.AttemptedAt); err != nil {
			continue
		}
		if uid.Valid {
			x.UserID = &uid.String
		}
		if ip.Valid {
			x.IPAddress = &ip.String
		}
		if ua.Valid {
			x.UserAgent = &ua.String
		}
		out = append(out, x)
	}
	writeJSON(w, 200, out)
}
func (a *App) sessions(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	rows, err := a.db.QueryContext(r.Context(), `SELECT id,device_hint,ip_address,expires_at,is_revoked,created_at FROM session.refresh_token WHERE user_id=$1 ORDER BY created_at DESC LIMIT 30`, u.ID)
	if err != nil {
		writeJSON(w, 500, apiError{Error: "no se pudieron consultar sesiones"})
		return
	}
	defer rows.Close()
	type row struct {
		ID                    string `json:"id"`
		DeviceHint, IPAddress *string
		ExpiresAt             time.Time
		IsRevoked             bool
		CreatedAt             time.Time
	}
	out := []row{}
	for rows.Next() {
		var x row
		var d, ip sql.NullString
		if err := rows.Scan(&x.ID, &d, &ip, &x.ExpiresAt, &x.IsRevoked, &x.CreatedAt); err != nil {
			continue
		}
		if d.Valid {
			x.DeviceHint = &d.String
		}
		if ip.Valid {
			x.IPAddress = &ip.String
		}
		out = append(out, x)
	}
	writeJSON(w, 200, out)
}
