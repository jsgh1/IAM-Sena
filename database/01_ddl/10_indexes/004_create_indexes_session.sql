
CREATE INDEX ix_refresh_token_user_id_is_revoked
    ON session.refresh_token (user_id, is_revoked);