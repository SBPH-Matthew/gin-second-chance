package requests

type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=2"`
	Price       float64 `json:"price" validate:"required,min=0"`
	Description string  `json:"description"`
}
