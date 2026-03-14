UPDATE accounts
SET "name"       = :name,
    "type"       = :type,
    "enabled"    = :enabled,
    "updated_at" = :updated_at
WHERE id = :id
