package requests

type RegisterRequest struct {
	FirstName       string `json:"first_name" validate:"required,min=2"`
	LastName        string `json:"last_name" validate:"required,min=2"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=6,eqfield=Password"`
}
