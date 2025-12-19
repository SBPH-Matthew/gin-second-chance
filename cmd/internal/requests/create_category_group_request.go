package requests

type CategoryGroupCreateRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}
