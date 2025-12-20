package requests

type CreateProductConditionRequest struct {
	Name string `json:"name" validate:"required;min=2;max=100"`
}
