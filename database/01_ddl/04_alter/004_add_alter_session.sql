
ALTER TABLE session.refresh_token
    ADD CONSTRAINT fk_refresh_token_user
    FOREIGN KEY (user_id)
    REFERENCES identity.user (id)
    ON UPDATE RESTRICT
    ON DELETE CASCADE;

ALTER TABLE session.password_reset_request
    ADD CONSTRAINT fk_password_reset_request_user
    FOREIGN KEY (user_id)
    REFERENCES identity.user (id)
    ON UPDATE RESTRICT
    ON DELETE CASCADE;