
CREATE INDEX ix_audit_login_email_attempted_attempted_at
    ON identity_audit.audit_login (email_attempted, attempted_at);