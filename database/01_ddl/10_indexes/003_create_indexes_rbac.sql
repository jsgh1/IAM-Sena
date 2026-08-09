
CREATE INDEX ix_role_feature_feature_id
    ON rbac.role_feature (feature_id);

CREATE INDEX ix_user_role_user_id
    ON rbac.user_role (user_id);

CREATE INDEX ix_user_scope_override_user_id_feature_id
    ON rbac.user_scope_override (user_id, feature_id);