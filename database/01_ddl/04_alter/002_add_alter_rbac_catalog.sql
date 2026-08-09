
ALTER TABLE rbac_catalog.feature
    ADD CONSTRAINT fk_feature_module
    FOREIGN KEY (module_id)
    REFERENCES rbac_catalog.module (id)
    ON UPDATE RESTRICT
    ON DELETE RESTRICT;