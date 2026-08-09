
CREATE TABLE rbac.role (
    id              UUID          NOT NULL DEFAULT gen_random_uuid(),
    name            VARCHAR(50)   NOT NULL,
    display_name    VARCHAR(100)  NOT NULL,
    description     TEXT,
    is_system_role  BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT pk_role
        PRIMARY KEY (id),

    CONSTRAINT uq_role_name
        UNIQUE (name)
);

CREATE TABLE rbac.role_feature (
    id          UUID          NOT NULL DEFAULT gen_random_uuid(),
    role_id     UUID          NOT NULL,
    feature_id  UUID          NOT NULL,
    scope_type  VARCHAR(30)   NOT NULL,

    CONSTRAINT pk_role_feature
        PRIMARY KEY (id),

    CONSTRAINT uq_role_feature_role_id_feature_id
        UNIQUE (role_id, feature_id),

    CONSTRAINT ck_role_feature_scope_type
        CHECK (scope_type IN ('GLOBAL', 'TRAINING_CENTER', 'AREA', 'OWN_FICHAS', 'OWN_SCHEDULE', 'OWN_PROFILE', 'OWN_FICHA_AS_LEARNER'))
);

CREATE TABLE rbac.user_role (
    id                    UUID          NOT NULL DEFAULT gen_random_uuid(),
    user_id               UUID          NOT NULL,
    role_id               UUID          NOT NULL,
    training_center_id    UUID,
    assigned_by           UUID          NOT NULL,
    assigned_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    expires_at            TIMESTAMPTZ,

    CONSTRAINT pk_user_role
        PRIMARY KEY (id),

    CONSTRAINT uq_user_role_user_id_role_id_training_center_id
        UNIQUE (user_id, role_id, training_center_id)
);

CREATE TABLE rbac.user_scope_override (
    id           UUID          NOT NULL DEFAULT gen_random_uuid(),
    user_id      UUID          NOT NULL,
    feature_id   UUID          NOT NULL,
    scope_type   VARCHAR(30)   NOT NULL,
    is_allowed   BOOLEAN       NOT NULL DEFAULT TRUE,
    reason       TEXT          NOT NULL,
    granted_by   UUID          NOT NULL,
    expires_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT pk_user_scope_override
        PRIMARY KEY (id),

    CONSTRAINT ck_user_scope_override_scope_type
        CHECK (scope_type IN ('GLOBAL', 'TRAINING_CENTER', 'AREA', 'OWN_FICHAS', 'OWN_SCHEDULE', 'OWN_PROFILE', 'OWN_FICHA_AS_LEARNER'))
);