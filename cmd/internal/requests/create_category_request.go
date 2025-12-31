package requests

type CreateCategoryRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=100"`
	CategoryGroup uint   `json:"category_group" validate:"required,min=1,max=100"`
	Status        uint   `json:"status" validate:"required,min=1,max=100"`
}
