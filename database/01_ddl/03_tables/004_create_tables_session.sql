
CREATE TABLE session.refresh_token (
    id            UUID          NOT NULL DEFAULT gen_random_uuid(),
    user_id       UUID          NOT NULL,
    token_hash    TEXT          NOT NULL,
    device_hint   VARCHAR(200),
    ip_address    VARCHAR(45),
    expires_at    TIMESTAMPTZ   NOT NULL DEFAULT (now() + INTERVAL '7 days'),
    is_revoked    BOOLEAN       NOT NULL DEFAULT FALSE,
    revoked_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT pk_refresh_token
        PRIMARY KEY (id),

    CONSTRAINT uq_refresh_token_token_hash
        UNIQUE (token_hash)
);

CREATE TABLE session.password_reset_request (
    id            UUID          NOT NULL DEFAULT gen_random_uuid(),
    user_id       UUID          NOT NULL,
    token_hash    TEXT          NOT NULL,
    expires_at    TIMESTAMPTZ   NOT NULL DEFAULT (now() + INTERVAL '1 hour'),
    is_used       BOOLEAN       NOT NULL DEFAULT FALSE,
    requested_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    ip_address    VARCHAR(45),

    CONSTRAINT pk_password_reset_request
        PRIMARY KEY (id)
);