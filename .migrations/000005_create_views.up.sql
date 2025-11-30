-- Description: Add users roles view.
CREATE OR REPLACE VIEW view_user_roles AS
SELECT u.id AS user_id,
       u.name,
       u.email,
       r.name AS user_role
FROM users AS u
JOIN roles AS r ON u.role_id = r.id
