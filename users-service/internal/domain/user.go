package domain

type User struct {
	ID        uint   `json:"id"`
	Username  string `json:"usermame"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
