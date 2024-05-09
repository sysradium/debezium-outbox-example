package domain

import "github.com/google/uuid"

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"usermame"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}
