
ALTER TABLE identity_audit.audit_login
    ADD CONSTRAINT fk_audit_login_user
    FOREIGN KEY (user_id)
    REFERENCES identity.user (id)
    ON UPDATE RESTRICT
    ON DELETE SET NULL;