package eventbus

type UserRegisteredEvent struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

func NewUserRegisteredEvent(userID, producer, traceID, email, displayName, avatarURL string) *Envelope[UserRegisteredEvent] {
	return NewEnvelope(
		EventTypeUserRegistered,
		1,
		producer,
		traceID,
		UserRegisteredEvent{
			UserID:      userID,
			Email:       email,
			DisplayName: displayName,
			AvatarURL:   avatarURL,
		},
	)
}

type UserUpdatedFields struct {
	Email       *string `json:"email,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

type UserUpdatedEvent struct {
	UserID  string            `json:"user_id"`
	Changes UserUpdatedFields `json:"changes"`
	Reason  string            `json:"reason,omitempty"`
}

func NewUserUpdatedEvent(userID, producer, traceID string, changes UserUpdatedFields, reason string) *Envelope[UserUpdatedEvent] {
	return NewEnvelope(
		EventTypeUserUpdated,
		1,
		producer,
		traceID,
		UserUpdatedEvent{
			UserID:  userID,
			Changes: changes,
			Reason:  reason,
		},
	)
}

type UserDeletedEvent struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason,omitempty"`
}

func NewUserDeletedEvent(userID, producer, traceID, reason string) *Envelope[UserDeletedEvent] {
	return NewEnvelope(
		EventTypeUserDeleted,
		1,
		producer,
		traceID,
		UserDeletedEvent{
			UserID: userID,
			Reason: reason,
		},
	)
}
