
CREATE TABLE identity_audit.audit_login (
    id                UUID          NOT NULL DEFAULT gen_random_uuid(),
    user_id           UUID,
    email_attempted   VARCHAR(255)  NOT NULL,
    outcome           VARCHAR(20)   NOT NULL,
    ip_address        VARCHAR(45),
    user_agent        TEXT,
    attempted_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT pk_audit_login
        PRIMARY KEY (id),

    CONSTRAINT ck_audit_login_outcome
        CHECK (outcome IN ('SUCCESS', 'INVALID_PASSWORD', 'USER_NOT_FOUND', 'ACCOUNT_LOCKED', 'TOKEN_EXPIRED'))
);