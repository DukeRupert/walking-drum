package dto

// ProductRequest represents the request to create or update a product
type ProductRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Origin      string `json:"origin"`
	RoastLevel  string `json:"roast_level"`
}

// ProductPriceRequest represents the request to create or update a product price
type ProductPriceRequest struct {
	Weight     string  `json:"weight" validate:"required"`
	Grind      string  `json:"grind" validate:"required"`
	Price      float64 `json:"price" validate:"required,gt=0"`
	IsDefault  bool    `json:"is_default"`
}

// CustomerRequest represents the request to register or update a customer
type CustomerRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required"`
	Phone string `json:"phone"`
}

// AddressRequest represents the request to create or update an address
type AddressRequest struct {
	Line1      string `json:"line1" validate:"required"`
	Line2      string `json:"line2"`
	City       string `json:"city" validate:"required"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code" validate:"required"`
	Country    string `json:"country" validate:"required,len=2"`
	IsDefault  bool   `json:"is_default"`
}

// SubscriptionRequest represents the request to create a subscription
type SubscriptionRequest struct {
	CustomerID   int64 `json:"customer_id" validate:"required"`
	PriceID      int64 `json:"price_id" validate:"required"`
	AddressID    int64 `json:"address_id" validate:"required"`
	PaymentToken string `json:"payment_token,omitempty"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code"`
}

// PaginatedResponse wraps paginated results
type PaginatedResponse struct {
	Total       int         `json:"total"`
	Page        int         `json:"page"`
	Limit       int         `json:"limit"`
	TotalPages  int         `json:"total_pages"`
	HasMore     bool        `json:"has_more"`
	Data        interface{} `json:"data"`
}