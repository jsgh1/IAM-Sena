-- DEMO - usuario SYSTEM_ADMIN inicial (docs: seed de iam) + su rol
INSERT INTO identity."user" (id, email, password_hash, first_name, last_name, actor_type) VALUES
  ('c0000000-0000-0000-0000-000000000001','admin@sena.edu.co','$2b$12$DEMOhashDEMOhashDEMOhashDEMOhashDEMOhashDE','System','Admin','USER')
ON CONFLICT (id) DO NOTHING;
INSERT INTO rbac.user_role (user_id, role_id, assigned_by)
SELECT 'c0000000-0000-0000-0000-000000000001', r.id, 'c0000000-0000-0000-0000-000000000001'
FROM rbac.role r WHERE r.name='SYSTEM_ADMIN'
ON CONFLICT DO NOTHING;
