package requests

type CreateVehicleTypeRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}
