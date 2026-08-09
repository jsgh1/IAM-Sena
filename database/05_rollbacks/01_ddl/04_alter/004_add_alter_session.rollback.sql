
ALTER TABLE session.password_reset_request
    DROP CONSTRAINT fk_password_reset_request_user;

ALTER TABLE session.refresh_token
    DROP CONSTRAINT fk_refresh_token_user;