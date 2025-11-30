UPDATE
    users
SET
    "role_id"       = :role_id,
    "name"          = :name,
    "email"         = :email,
    "password_hash" = :password_hash,
    "department"    = :department,
    "enabled"       = :enabled,
    "date_updated"  = :date_updated
WHERE id = :id