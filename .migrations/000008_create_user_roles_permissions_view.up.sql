CREATE OR REPLACE VIEW vw_user_roles_permissions AS
SELECT
    u.id    AS user_id,
    u.name  AS user_name,
    u.email AS user_email,
    r.id    AS role_id,
    r.name  AS role_name,
    p.id    AS permission_id,
    p.name  AS permission_name
FROM users AS u
JOIN user_roles AS ur ON ur.user_id = u.id
JOIN roles AS r ON r.id = ur.role_id
LEFT JOIN role_permissions AS rp ON rp.role_id = r.id
LEFT JOIN permissions AS p ON p.id = rp.permission_id
