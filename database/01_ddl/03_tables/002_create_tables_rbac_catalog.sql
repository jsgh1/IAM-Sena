
CREATE TABLE rbac_catalog.module (
    id             UUID          NOT NULL DEFAULT gen_random_uuid(),
    code           VARCHAR(30)   NOT NULL,
    name           VARCHAR(100)  NOT NULL,
    description    TEXT,
    display_order  SMALLINT      NOT NULL,
    icon_key       VARCHAR(50),
    is_active      BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT pk_module
        PRIMARY KEY (id),

    CONSTRAINT uq_module_code
        UNIQUE (code)
);

CREATE TABLE rbac_catalog.feature (
    id             UUID          NOT NULL DEFAULT gen_random_uuid(),
    module_id      UUID          NOT NULL,
    code           VARCHAR(60)   NOT NULL,
    name           VARCHAR(120)  NOT NULL,
    description    TEXT,
    action_level   VARCHAR(20)   NOT NULL,
    is_active      BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT pk_feature
        PRIMARY KEY (id),

    CONSTRAINT uq_feature_code
        UNIQUE (code),

    CONSTRAINT ck_feature_action_level
        CHECK (action_level IN ('READ', 'WRITE', 'DELETE', 'PUBLISH', 'APPROVE'))
);