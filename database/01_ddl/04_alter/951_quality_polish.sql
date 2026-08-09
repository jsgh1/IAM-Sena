-- Quality polish v2: unicidad de email case-insensitive (evita cuentas duplicadas por mayusculas)
CREATE UNIQUE INDEX IF NOT EXISTS uq_user_email_lower ON identity."user" (lower(email));
