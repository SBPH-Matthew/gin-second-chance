package requests

type UpdateUserPasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}
