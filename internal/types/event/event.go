// Package event defines domain event type and topic constants.
package event

const (
	UserCreated = "user.created"
	UserUpdated = "user.updated"
	UserDeleted = "user.deleted"

	UserTopic = "user-events"
)
