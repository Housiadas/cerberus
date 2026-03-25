UPDATE email_notification_outbox
SET retry_count = retry_count + 1
WHERE id IN (:ids)
