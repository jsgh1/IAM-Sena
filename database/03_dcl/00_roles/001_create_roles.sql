-- Roles de aplicacion (least-privilege) para el dominio iam
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='iam_reader') THEN CREATE ROLE iam_reader NOLOGIN; END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='iam_writer') THEN CREATE ROLE iam_writer NOLOGIN; END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='iam_admin')  THEN CREATE ROLE iam_admin  NOLOGIN; END IF;
END $$;
