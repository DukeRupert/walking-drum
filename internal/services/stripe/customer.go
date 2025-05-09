package stripe

import (
	"context"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v82"
)

// CustomerCreateParams defines parameters for creating a customer in Stripe
type CustomerCreateParams struct {
    Email       string            `json:"email"`
    Name        string            `json:"name"`
    Phone       string            `json:"phone,omitempty"`
    Description string            `json:"description,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

// CreateCustomer creates a new customer in Stripe
func (c *Client) CreateCustomer(ctx context.Context, params *CustomerCreateParams) (*stripe.Customer, error) {
    // Log the function entry
    c.logger.Info().
        Str("email", params.Email).
        Str("name", params.Name).
        Msg("Executing CreateCustomer()")

    // Check if client is initialized
    if c.api == nil {
        c.logger.Error().Msg("Stripe client not initialized")
        return nil, errors.New("stripe client not initialized")
    }

    // Create the Stripe params
    customerParams := &stripe.CustomerCreateParams{
        Email: stripe.String(params.Email),
    }

    // Log the base parameters
    c.logger.Info().
        Str("email", *customerParams.Email).
        Msg("Base customer parameters")

    // Add name if provided
    if params.Name != "" {
        customerParams.Name = stripe.String(params.Name)
        c.logger.Info().
            Str("name", params.Name).
            Msg("Adding name to customer parameters")
    }

    // Add phone if provided
    if params.Phone != "" {
        customerParams.Phone = stripe.String(params.Phone)
        c.logger.Info().
            Str("phone", params.Phone).
            Msg("Adding phone to customer parameters")
    }

    // Add description if provided
    if params.Description != "" {
        customerParams.Description = stripe.String(params.Description)
        c.logger.Info().
            Str("description", params.Description).
            Msg("Adding description to customer parameters")
    }

    // Add metadata if provided
    if len(params.Metadata) > 0 {
        customerParams.Metadata = make(map[string]string)
        for k, v := range params.Metadata {
            customerParams.Metadata[k] = v
        }

        // Log each metadata key-value pair
        for k, v := range params.Metadata {
            c.logger.Info().
                Str("key", k).
                Str("value", v).
                Msg("Adding metadata to customer parameters")
        }
    }

    // Log that we're about to call the Stripe API
    c.logger.Info().Msg("About to call Stripe API to create customer")

    // Make the API call
    customer, err := c.api.V1Customers.Create(ctx, customerParams)

    // Log the result
    if err != nil {
        c.logger.Error().
            Err(err).
            Msg("Failed to create customer in Stripe")
        return nil, fmt.Errorf("failed to create customer in Stripe: %w", err)
    }

    // Log successful creation with Stripe ID
    c.logger.Info().
        Str("stripe_id", customer.ID).
        Str("email", customer.Email).
        Msg("Successfully created customer in Stripe")

    return customer, nil
}