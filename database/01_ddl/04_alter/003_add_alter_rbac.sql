
ALTER TABLE rbac.role_feature
    ADD CONSTRAINT fk_role_feature_role
    FOREIGN KEY (role_id)
    REFERENCES rbac.role (id)
    ON UPDATE RESTRICT
    ON DELETE CASCADE;

ALTER TABLE rbac.role_feature
    ADD CONSTRAINT fk_role_feature_feature
    FOREIGN KEY (feature_id)
    REFERENCES rbac_catalog.feature (id)
    ON UPDATE RESTRICT
    ON DELETE RESTRICT;

ALTER TABLE rbac.user_role
    ADD CONSTRAINT fk_user_role_user
    FOREIGN KEY (user_id)
    REFERENCES identity.user (id)
    ON UPDATE RESTRICT
    ON DELETE CASCADE;

ALTER TABLE rbac.user_role
    ADD CONSTRAINT fk_user_role_role
    FOREIGN KEY (role_id)
    REFERENCES rbac.role (id)
    ON UPDATE RESTRICT
    ON DELETE RESTRICT;

ALTER TABLE rbac.user_role
    ADD CONSTRAINT fk_user_role_assigned_by
    FOREIGN KEY (assigned_by)
    REFERENCES identity.user (id)
    ON UPDATE RESTRICT
    ON DELETE RESTRICT;

ALTER TABLE rbac.user_scope_override
    ADD CONSTRAINT fk_user_scope_override_user
    FOREIGN KEY (user_id)
    REFERENCES identity.user (id)
    ON UPDATE RESTRICT
    ON DELETE CASCADE;

ALTER TABLE rbac.user_scope_override
    ADD CONSTRAINT fk_user_scope_override_feature
    FOREIGN KEY (feature_id)
    REFERENCES rbac_catalog.feature (id)
    ON UPDATE RESTRICT
    ON DELETE RESTRICT;

ALTER TABLE rbac.user_scope_override
    ADD CONSTRAINT fk_user_scope_override_granted_by
    FOREIGN KEY (granted_by)
    REFERENCES identity.user (id)
    ON UPDATE RESTRICT
    ON DELETE RESTRICT;