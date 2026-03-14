UPDATE
    users
SET
    "name"          = :name,
    "email"         = :email,
    "password_hash" = :password_hash,
    "department"    = :department,
    "account_id"    = :account_id,
    "enabled"       = :enabled,
    "updated_at"  = :updated_at
WHERE id = :id
