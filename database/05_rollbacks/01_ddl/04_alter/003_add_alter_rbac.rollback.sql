
ALTER TABLE rbac.user_scope_override
    DROP CONSTRAINT fk_user_scope_override_granted_by;

ALTER TABLE rbac.user_scope_override
    DROP CONSTRAINT fk_user_scope_override_feature;

ALTER TABLE rbac.user_scope_override
    DROP CONSTRAINT fk_user_scope_override_user;

ALTER TABLE rbac.user_role
    DROP CONSTRAINT fk_user_role_assigned_by;

ALTER TABLE rbac.user_role
    DROP CONSTRAINT fk_user_role_role;

ALTER TABLE rbac.user_role
    DROP CONSTRAINT fk_user_role_user;

ALTER TABLE rbac.role_feature
    DROP CONSTRAINT fk_role_feature_feature;

ALTER TABLE rbac.role_feature
    DROP CONSTRAINT fk_role_feature_role;