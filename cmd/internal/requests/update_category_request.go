package requests

type UpdateCategoryRequest struct {
	Name   string `json:"name" validate:"required,min=2,max=100"`
	Status uint   `json:"status" validate:"required,min=1"`
}
