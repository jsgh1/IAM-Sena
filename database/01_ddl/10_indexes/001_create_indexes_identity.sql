
CREATE INDEX ix_user_actor_id
    ON identity.user (actor_id);

CREATE INDEX ix_user_is_active
    ON identity.user (is_active)
    WHERE is_active = TRUE;

CREATE INDEX ix_user_last_name_first_name
    ON identity.user (last_name, first_name);