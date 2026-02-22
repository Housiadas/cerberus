UPDATE outbox
SET processed_at = :processed_at
WHERE id IN (:ids)
