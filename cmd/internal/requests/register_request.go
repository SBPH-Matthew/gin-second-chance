package requests

type RegisterRequest struct {
	FirstName       string `json:"first_name" binding:"required,min=2"`
	LastName        string `json:"last_name" binding:"required,min=2"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=6,eqfield=Password"`
}
