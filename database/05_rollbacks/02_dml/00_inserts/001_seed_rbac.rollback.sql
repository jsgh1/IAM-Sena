-- Rollback del seed RBAC
DELETE FROM rbac.role_feature;
DELETE FROM rbac_catalog.feature;
DELETE FROM rbac.role;
DELETE FROM rbac_catalog.module;
