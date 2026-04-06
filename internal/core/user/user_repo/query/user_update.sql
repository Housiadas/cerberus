UPDATE
    users
SET
    "name"          = :name,
    "email"         = :email,
    "password_hash" = :password_hash,
    "department"    = :department,
    "enabled"       = :enabled,
    "account_id"    = :account_id,
    "updated_at"  = :updated_at
WHERE id = :id
