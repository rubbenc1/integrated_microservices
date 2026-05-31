package dto

type UserCreatedEvent struct {
	ID   	 string `json:"id"`
	Email    string `json:"email"`
	UserName string `json:"username"`
}
