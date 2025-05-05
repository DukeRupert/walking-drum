// internal/domain/dto/customer_dto.go
package dto

import (
	"context"
	"regexp"
)

// CustomerCreateDTO represents the data needed to create a new customer
type CustomerCreateDTO struct {
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// Valid validates the CustomerCreateDTO
func (c *CustomerCreateDTO) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if c.Email == "" {
		problems["email"] = "email is required"
	} else {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(c.Email) {
			problems["email"] = "invalid email format"
		}
	}

	if c.FirstName == "" {
		problems["first_name"] = "first name is required"
	}

	if c.LastName == "" {
		problems["last_name"] = "last name is required"
	}

	if c.PhoneNumber != "" {
		phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
		if !phoneRegex.MatchString(c.PhoneNumber) {
			problems["phone_number"] = "invalid phone number format"
		}
	}

	return problems
}

// CustomerUpdateDTO represents the data that can be updated for a customer
type CustomerUpdateDTO struct {
	Email       string `json:"email,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Active      *bool  `json:"active,omitempty"`
}

// Valid validates the CustomerUpdateDTO
func (c *CustomerUpdateDTO) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if c.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(c.Email) {
			problems["email"] = "invalid email format"
		}
	}

	if c.PhoneNumber != "" {
		phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
		if !phoneRegex.MatchString(c.PhoneNumber) {
			problems["phone_number"] = "invalid phone number format"
		}
	}

	return problems
}

// CustomerResponseDTO represents the data returned to the client
type CustomerResponseDTO struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Active      bool   `json:"active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// AddressCreateDTO represents the data needed to create a new address
type AddressCreateDTO struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	IsDefault  bool   `json:"is_default"`
}

// Valid validates the AddressCreateDTO
func (a *AddressCreateDTO) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if a.Line1 == "" {
		problems["line1"] = "address line 1 is required"
	}

	if a.City == "" {
		problems["city"] = "city is required"
	}

	if a.State == "" {
		problems["state"] = "state is required"
	}

	if a.PostalCode == "" {
		problems["postal_code"] = "postal code is required"
	}

	if a.Country == "" {
		problems["country"] = "country is required"
	}

	return problems
}

// AddressUpdateDTO represents the data that can be updated for an address
type AddressUpdateDTO struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
	IsDefault  *bool  `json:"is_default,omitempty"`
}

// Valid validates the AddressUpdateDTO
func (a *AddressUpdateDTO) Valid(ctx context.Context) map[string]string {
	// All fields are optional for update, so no specific validation
	return make(map[string]string)
}

// AddressResponseDTO represents the data returned to the client
type AddressResponseDTO struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	IsDefault  bool   `json:"is_default"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}