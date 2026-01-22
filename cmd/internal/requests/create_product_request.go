package requests

type CreateProductRequest struct {
	Name        string  `json:"name" form:"name" validate:"required,min=2"`
	Price       float64 `json:"price" form:"price" validate:"required,min=0"`
	Description string  `json:"description" form:"description"`
	Location    string  `json:"location" form:"location" validate:"required"`
	Status      string  `json:"status" form:"status" validate:"required"`
	Condition   string  `json:"condition" form:"condition" validate:"required"`
	Category    string  `json:"category" form:"category" validate:"required"`
}
