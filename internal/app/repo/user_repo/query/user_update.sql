UPDATE
    users
SET
    "name"          = :name,
    "email"         = :email,
    "password_hash" = :password_hash,
    "department"    = :department,
    "enabled"       = :enabled,
    "updated_at"  = :updated_at
WHERE id = :id
