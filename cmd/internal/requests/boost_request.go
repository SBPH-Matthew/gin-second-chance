package requests

type CreateBoostRequest struct {
	ItemType      string `json:"item_type" validate:"required,oneof=product vehicle"`
	ItemID        uint   `json:"item_id" validate:"required,min=1"`
	BoostType     string `json:"boost_type" validate:"required,oneof=premium featured top"`
	DurationHours int    `json:"duration_hours" validate:"required,min=1,max=168"`
}

type UpdateBoostRequest struct {
	BoostType     string `json:"boost_type" validate:"required,oneof=premium featured top"`
	DurationHours int    `json:"duration_hours" validate:"required,min=1,max=168"`
}

type ProcessPaymentRequest struct {
	PaymentMethod string `json:"payment_method" validate:"required"`
	// Add payment provider specific fields as needed
}
