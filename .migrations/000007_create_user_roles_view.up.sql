CREATE OR REPLACE VIEW vw_user_roles AS
SELECT
    u.id    AS user_id,
    u.name  AS user_name,
    u.email AS user_email,
    r.id    AS role_id,
    r.name  AS role_name,
FROM users AS u
JOIN user_roles AS ur ON ur.user_id = u.id
JOIN roles AS r ON r.id = ur.role_id;
