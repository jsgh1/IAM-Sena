
CREATE TABLE identity.user (
    id              UUID          NOT NULL DEFAULT gen_random_uuid(),
    email           VARCHAR(255)  NOT NULL,
    password_hash   TEXT          NOT NULL,
    first_name      VARCHAR(100)  NOT NULL,
    last_name       VARCHAR(100)  NOT NULL,
    actor_type      VARCHAR(20)   NOT NULL,
    actor_id        UUID,
    is_active       BOOLEAN       NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    failed_attempts SMALLINT      NOT NULL DEFAULT 0,
    locked_until    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT pk_user
        PRIMARY KEY (id),

    CONSTRAINT uq_user_email
        UNIQUE (email),

    CONSTRAINT ck_user_actor_type
        CHECK (actor_type IN ('USER', 'INSTRUCTOR', 'LEARNER'))
);