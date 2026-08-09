-- Quality fixes (v2.0.0): id DEFAULT gen_random_uuid() e indices de FK faltantes
CREATE INDEX IF NOT EXISTS ix_audit_login_user_id ON identity_audit.audit_login (user_id);
CREATE INDEX IF NOT EXISTS ix_feature_module_id ON rbac_catalog.feature (module_id);
CREATE INDEX IF NOT EXISTS ix_user_role_assigned_by ON rbac.user_role (assigned_by);
CREATE INDEX IF NOT EXISTS ix_user_role_role_id ON rbac.user_role (role_id);
CREATE INDEX IF NOT EXISTS ix_user_scope_override_feature_id ON rbac.user_scope_override (feature_id);
CREATE INDEX IF NOT EXISTS ix_user_scope_override_granted_by ON rbac.user_scope_override (granted_by);
CREATE INDEX IF NOT EXISTS ix_password_reset_request_user_id ON session.password_reset_request (user_id);
