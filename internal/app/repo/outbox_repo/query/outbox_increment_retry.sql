UPDATE outbox
SET retry_count = retry_count + 1
WHERE id IN (:ids)
