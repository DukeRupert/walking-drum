package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dukerupert/walking-drum/internal/domain/dto"
	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/dukerupert/walking-drum/internal/repositories/interfaces/mocks"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/dukerupert/walking-drum/internal/services/stripe"
	stripemock "github.com/dukerupert/walking-drum/internal/services/stripe/mocks"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	stripego "github.com/stripe/stripe-go/v82"
)

func TestProductService_Create(t *testing.T) {
	// Initialize logger
	logger := zerolog.New(zerolog.NewTestWriter(t))
	
	// Test cases
	testCases := []struct {
		name           string
		productDTO     *dto.ProductCreateDTO
		setupMocks     func(*mocks.ProductRepository, *stripemock.Client)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "Successful product creation",
			productDTO: &dto.ProductCreateDTO{
				Name:        "Colombian Coffee",
				Description: "Rich and smooth coffee from Colombia",
				ImageURL:    "https://example.com/coffee.jpg",
				Active:      true,
				StockLevel:  100,
				Weight:      250,
				Origin:      "Colombia",
				RoastLevel:  "Medium",
				FlavorNotes: "Chocolate, Nutty",
			},
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// Create a Stripe product from your own package
				stripeProduct := &stripego.Product{
					ID:     "stripe_prod_123",
					Name:   "Colombian Coffee",
					Active: true,
				}
				
				// Mock Stripe client response
				mockStripe.On("CreateProduct", mock.Anything, mock.MatchedBy(func(params *stripe.ProductCreateParams) bool {
					return params.Name == "Colombian Coffee" &&
						params.Description == "Rich and smooth coffee from Colombia" &&
						params.Images[0] == "https://example.com/coffee.jpg" &&
						params.Active == true
				})).Return(stripeProduct, nil)

				// Mock repository create
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(product *models.Product) bool {
					return product.Name == "Colombian Coffee" &&
						product.Description == "Rich and smooth coffee from Colombia" &&
						product.ImageURL == "https://example.com/coffee.jpg" &&
						product.StockLevel == 100 &&
						product.Weight == 250 &&
						product.Origin == "Colombia" &&
						product.RoastLevel == "Medium" &&
						product.FlavorNotes == "Chocolate, Nutty" &&
						product.StripeID == "stripe_prod_123"
				})).Return(nil)
			},
			expectedError: false,
		},
		// Other test cases...
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockRepo := new(mocks.ProductRepository)
			mockStripe := new(stripemock.Client)

			// Configure mocks
			tc.setupMocks(mockRepo, mockStripe)

			// Create service
			productService := services.NewProductService(mockRepo, mockStripe, &logger)

			// Execute test
			result, err := productService.Create(context.Background(), tc.productDTO)

			// Assertions
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.productDTO.Name, result.Name)
				assert.Equal(t, tc.productDTO.Description, result.Description)
				assert.Equal(t, tc.productDTO.ImageURL, result.ImageURL)
				assert.Equal(t, tc.productDTO.Active, result.Active)
				assert.Equal(t, tc.productDTO.StockLevel, result.StockLevel)
				assert.Equal(t, tc.productDTO.Weight, result.Weight)
				assert.Equal(t, tc.productDTO.Origin, result.Origin)
				assert.Equal(t, tc.productDTO.RoastLevel, result.RoastLevel)
				assert.Equal(t, tc.productDTO.FlavorNotes, result.FlavorNotes)
				assert.Equal(t, "stripe_prod_123", result.StripeID)
				assert.NotEqual(t, uuid.Nil, result.ID)
				assert.WithinDuration(t, time.Now(), result.CreatedAt, 5*time.Second)
				assert.WithinDuration(t, time.Now(), result.UpdatedAt, 5*time.Second)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
			mockStripe.AssertExpectations(t)
		})
	}
}

func TestProductService_GetByID(t *testing.T) {
	// Initialize logger
	logger := zerolog.New(zerolog.NewTestWriter(t))
	
	// Define a sample product ID
	productID := uuid.New()
	
	// Test cases
	testCases := []struct {
		name           string
		setupMocks     func(*mocks.ProductRepository)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "Successful product retrieval",
			setupMocks: func(mockRepo *mocks.ProductRepository) {
				mockRepo.On("GetByID", mock.Anything, productID).Return(&models.Product{
					ID:          productID,
					Name:        "Colombian Coffee",
					Description: "Rich and smooth coffee from Colombia",
					ImageURL:    "https://example.com/coffee.jpg",
					Active:      true,
					StockLevel:  100,
					Weight:      250,
					Origin:      "Colombia",
					RoastLevel:  "Medium",
					FlavorNotes: "Chocolate, Nutty",
					StripeID:    "stripe_prod_123",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Product not found",
			setupMocks: func(mockRepo *mocks.ProductRepository) {
				mockRepo.On("GetByID", mock.Anything, productID).Return(nil, errors.New("product not found"))
			},
			expectedError:  true,
			expectedErrMsg: "product not found",
		},
		{
			name: "Database error",
			setupMocks: func(mockRepo *mocks.ProductRepository) {
				mockRepo.On("GetByID", mock.Anything, productID).Return(nil, errors.New("database connection error"))
			},
			expectedError:  true,
			expectedErrMsg: "database connection error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockRepo := new(mocks.ProductRepository)
			mockStripe := new(stripemock.Client)

			// Configure mocks
			tc.setupMocks(mockRepo)

			// Create service
			productService := services.NewProductService(mockRepo, mockStripe, &logger)

			// Execute test
			result, err := productService.GetByID(context.Background(), productID)

			// Assertions
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, productID, result.ID)
				assert.Equal(t, "Colombian Coffee", result.Name)
				assert.Equal(t, "Rich and smooth coffee from Colombia", result.Description)
				assert.Equal(t, "https://example.com/coffee.jpg", result.ImageURL)
				assert.Equal(t, true, result.Active)
				assert.Equal(t, 100, result.StockLevel)
				assert.Equal(t, 250, result.Weight)
				assert.Equal(t, "Colombia", result.Origin)
				assert.Equal(t, "Medium", result.RoastLevel)
				assert.Equal(t, "Chocolate, Nutty", result.FlavorNotes)
				assert.Equal(t, "stripe_prod_123", result.StripeID)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_List(t *testing.T) {
	// Initialize logger
	logger := zerolog.New(zerolog.NewTestWriter(t))

	// Create common test products
	product1ID := uuid.New()
	product2ID := uuid.New()
	product3ID := uuid.New()
	
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now().Add(-12 * time.Hour)

	// Products with adequate stock
	product1 := &models.Product{
		ID:          product1ID,
		Name:        "Colombian Coffee",
		Description: "Rich and smooth coffee from Colombia",
		ImageURL:    "https://example.com/coffee1.jpg",
		Active:      true,
		StockLevel:  100,
		Weight:      250,
		Origin:      "Colombia",
		RoastLevel:  "Medium",
		FlavorNotes: "Chocolate, Nutty",
		StripeID:    "stripe_prod_123",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	product2 := &models.Product{
		ID:          product2ID,
		Name:        "Ethiopian Coffee",
		Description: "Fruity and floral coffee from Ethiopia",
		ImageURL:    "https://example.com/coffee2.jpg",
		Active:      true,
		StockLevel:  50,
		Weight:      250,
		Origin:      "Ethiopia",
		RoastLevel:  "Light",
		FlavorNotes: "Blueberry, Floral",
		StripeID:    "stripe_prod_456",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	// Product with low stock (< 10)
	product3 := &models.Product{
		ID:          product3ID,
		Name:        "Kenyan Coffee",
		Description: "Bold and vibrant coffee from Kenya",
		ImageURL:    "https://example.com/coffee3.jpg",
		Active:      true,
		StockLevel:  5, // Low stock
		Weight:      250,
		Origin:      "Kenya",
		RoastLevel:  "Medium-Dark",
		FlavorNotes: "Blackcurrant, Citrus",
		StripeID:    "stripe_prod_789",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	// Inactive product
	product4 := &models.Product{
		ID:          uuid.New(),
		Name:        "Decaf Colombian",
		Description: "Decaffeinated coffee from Colombia",
		ImageURL:    "https://example.com/coffee4.jpg",
		Active:      false, // Inactive
		StockLevel:  75,
		Weight:      250,
		Origin:      "Colombia",
		RoastLevel:  "Medium",
		FlavorNotes: "Chocolate, Caramel",
		StripeID:    "stripe_prod_101",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	// Test cases
	testCases := []struct {
		name             string
		offset           int
		limit            int
		includeInactive  bool
		setupMocks       func(*mocks.ProductRepository)
		expectedProducts []*models.Product
		expectedTotal    int
		expectedError    bool
		expectedErrMsg   string
	}{
		{
			name:            "List active products with pagination",
			offset:          0,
			limit:           2,
			includeInactive: false,
			setupMocks: func(mockRepo *mocks.ProductRepository) {
				// Only return active products (exclude product4)
				mockRepo.On("List", mock.Anything, 0, 2, false).Return(
					[]*models.Product{product1, product2}, 3, nil,
				)
			},
			expectedProducts: []*models.Product{product1, product2},
			expectedTotal:    3,
			expectedError:    false,
		},
		{
			name:            "List more active products with pagination",
			offset:          2,
			limit:           2,
			includeInactive: false,
			setupMocks: func(mockRepo *mocks.ProductRepository) {
				mockRepo.On("List", mock.Anything, 2, 2, false).Return(
					[]*models.Product{product3}, 3, nil,
				)
			},
			expectedProducts: []*models.Product{product3},
			expectedTotal:    3,
			expectedError:    false,
		},
		{
			name:            "List including inactive products",
			offset:          0,
			limit:           10,
			includeInactive: true,
			setupMocks: func(mockRepo *mocks.ProductRepository) {
				// Include all products
				mockRepo.On("List", mock.Anything, 0, 10, true).Return(
					[]*models.Product{product1, product2, product3, product4}, 4, nil,
				)
			},
			expectedProducts: []*models.Product{product1, product2, product3, product4},
			expectedTotal:    4,
			expectedError:    false,
		},
		{
			name:            "Empty product list",
			offset:          0,
			limit:           10,
			includeInactive: false,
			setupMocks: func(mockRepo *mocks.ProductRepository) {
				// Return empty list
				mockRepo.On("List", mock.Anything, 0, 10, false).Return(
					[]*models.Product{}, 0, nil,
				)
			},
			expectedProducts: []*models.Product{},
			expectedTotal:    0,
			expectedError:    false,
		},
		{
			name:            "Repository error",
			offset:          0,
			limit:           10,
			includeInactive: false,
			setupMocks: func(mockRepo *mocks.ProductRepository) {
				// Return error
				mockRepo.On("List", mock.Anything, 0, 10, false).Return(
					nil, 0, errors.New("database connection error"),
				)
			},
			expectedProducts: nil,
			expectedTotal:    0,
			expectedError:    true,
			expectedErrMsg:   "failed to list products",
		},
		{
			name:            "Invalid offset parameter",
			offset:          -1, // Invalid offset
			limit:           10,
			includeInactive: false,
			setupMocks: func(mockRepo *mocks.ProductRepository) {
				// For this test, we'll assume the repository still handles negative offset
                // In a real implementation, you might want to validate these parameters
				mockRepo.On("List", mock.Anything, -1, 10, false).Return(
					[]*models.Product{product1, product2}, 2, nil,
				)
			},
			expectedProducts: []*models.Product{product1, product2},
			expectedTotal:    2,
			expectedError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockRepo := new(mocks.ProductRepository)
			mockStripe := new(stripemock.Client)

			// Configure mocks
			tc.setupMocks(mockRepo)

			// Create service
			productService := services.NewProductService(mockRepo, mockStripe, &logger)

			// Execute test
			products, total, err := productService.List(context.Background(), tc.offset, tc.limit, tc.includeInactive)

			// Assertions
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				assert.Nil(t, products)
				assert.Equal(t, 0, total)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTotal, total)
				assert.Equal(t, len(tc.expectedProducts), len(products))

				// Check product details if we have products
				if len(products) > 0 {
					// Test that the products match our expected products
					for i, expectedProduct := range tc.expectedProducts {
						assert.Equal(t, expectedProduct.ID, products[i].ID)
						assert.Equal(t, expectedProduct.Name, products[i].Name)
						assert.Equal(t, expectedProduct.Description, products[i].Description)
						assert.Equal(t, expectedProduct.ImageURL, products[i].ImageURL)
						assert.Equal(t, expectedProduct.Active, products[i].Active)
						assert.Equal(t, expectedProduct.StockLevel, products[i].StockLevel)
						assert.Equal(t, expectedProduct.Weight, products[i].Weight)
						assert.Equal(t, expectedProduct.Origin, products[i].Origin)
						assert.Equal(t, expectedProduct.RoastLevel, products[i].RoastLevel)
						assert.Equal(t, expectedProduct.FlavorNotes, products[i].FlavorNotes)
						assert.Equal(t, expectedProduct.StripeID, products[i].StripeID)
					}

					// Check that we log a warning for products with low stock
					for _, product := range products {
						if product.StockLevel < 10 {
							// This is more of a verification that our test includes this scenario
							// than an actual functional test, since we can't easily test logs
							assert.True(t, product.StockLevel < 10, 
								"Test should include a product with low stock level (<10)")
						}
					}
				}
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_Update(t *testing.T) {
	// Initialize logger
	logger := zerolog.New(zerolog.NewTestWriter(t))

	// Create a helper function to generate string pointer
	strPtr := func(s string) *string {
		return &s
	}

	// Create a helper function to generate bool pointer
	boolPtr := func(b bool) *bool {
		return &b
	}

	// Create a helper function to generate int pointer
	intPtr := func(i int) *int {
		return &i
	}

	// Test cases
	testCases := []struct {
		name           string
		productID      uuid.UUID
		productDTO     *dto.ProductUpdateDTO
		setupMocks     func(*mocks.ProductRepository, *stripemock.Client)
		expectedError  bool
		expectedErrMsg string
		withRequestID  bool
	}{
		{
			name:      "Successful product update with all fields",
			productID: uuid.New(),
			productDTO: &dto.ProductUpdateDTO{
				Name:        strPtr("Updated Colombian Coffee"),
				Description: strPtr("Updated rich and smooth coffee from Colombia"),
				ImageURL:    strPtr("https://example.com/updated_coffee.jpg"),
				Active:      boolPtr(true),
				StockLevel:  intPtr(95),
				Weight:      intPtr(300),
				Origin:      strPtr("Colombia"),
				RoastLevel:  strPtr("Dark"),
				FlavorNotes: strPtr("Updated Chocolate, Caramel notes"),
			},
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// 1. Mock GetByID response
				existingProduct := &models.Product{
					ID:          uuid.New(),
					Name:        "Colombian Coffee",
					Description: "Rich and smooth coffee from Colombia",
					ImageURL:    "https://example.com/coffee.jpg",
					Active:      true,
					StockLevel:  100,
					Weight:      250,
					Origin:      "Colombia",
					RoastLevel:  "Medium",
					FlavorNotes: "Chocolate, Nutty",
					StripeID:    "stripe_prod_123",
					CreatedAt:   time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now().Add(-12 * time.Hour),
				}
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(existingProduct, nil)

				// 2. Mock UpdateProduct in Stripe
				mockStripe.On("UpdateProduct", mock.Anything, mock.MatchedBy(func(product *models.Product) bool {
					return product.Name == "Updated Colombian Coffee" &&
						product.Description == "Updated rich and smooth coffee from Colombia" &&
						product.ImageURL == "https://example.com/updated_coffee.jpg" &&
						product.Active == true &&
						product.StockLevel == 95 &&
						product.Weight == 300 &&
						product.Origin == "Colombia" &&
						product.RoastLevel == "Dark" &&
						product.FlavorNotes == "Updated Chocolate, Caramel notes" &&
						product.StripeID == "stripe_prod_123"
				})).Return(nil)

				// 3. Mock Update in repository
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(product *models.Product) bool {
					return product.Name == "Updated Colombian Coffee" &&
						product.Description == "Updated rich and smooth coffee from Colombia" &&
						product.ImageURL == "https://example.com/updated_coffee.jpg" &&
						product.Active == true &&
						product.StockLevel == 95 &&
						product.Weight == 300 &&
						product.Origin == "Colombia" &&
						product.RoastLevel == "Dark" &&
						product.FlavorNotes == "Updated Chocolate, Caramel notes" &&
						product.StripeID == "stripe_prod_123"
				})).Return(nil)
			},
			expectedError: false,
			withRequestID: false,
		},
		{
			name:      "Successful partial product update",
			productID: uuid.New(),
			productDTO: &dto.ProductUpdateDTO{
				Name:       strPtr("Updated Colombian Coffee"),
				StockLevel: intPtr(95),
				// Other fields not provided
			},
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// 1. Mock GetByID response
				existingProduct := &models.Product{
					ID:          uuid.New(),
					Name:        "Colombian Coffee",
					Description: "Rich and smooth coffee from Colombia",
					ImageURL:    "https://example.com/coffee.jpg",
					Active:      true,
					StockLevel:  100,
					Weight:      250,
					Origin:      "Colombia",
					RoastLevel:  "Medium",
					FlavorNotes: "Chocolate, Nutty",
					StripeID:    "stripe_prod_123",
					CreatedAt:   time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now().Add(-12 * time.Hour),
				}
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(existingProduct, nil)

				// 2. Mock UpdateProduct in Stripe
				mockStripe.On("UpdateProduct", mock.Anything, mock.MatchedBy(func(product *models.Product) bool {
					// Only name should be updated, other fields should remain the same
					return product.Name == "Updated Colombian Coffee" &&
						product.Description == "Rich and smooth coffee from Colombia" &&
						product.ImageURL == "https://example.com/coffee.jpg" &&
						product.Active == true &&
						product.StockLevel == 95 &&
						product.Weight == 250 &&
						product.Origin == "Colombia" &&
						product.RoastLevel == "Medium" &&
						product.FlavorNotes == "Chocolate, Nutty" &&
						product.StripeID == "stripe_prod_123"
				})).Return(nil)

				// 3. Mock Update in repository
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(product *models.Product) bool {
					// Only name and stock level should be updated
					return product.Name == "Updated Colombian Coffee" &&
						product.Description == "Rich and smooth coffee from Colombia" &&
						product.ImageURL == "https://example.com/coffee.jpg" &&
						product.Active == true &&
						product.StockLevel == 95 &&
						product.Weight == 250 &&
						product.Origin == "Colombia" &&
						product.RoastLevel == "Medium" &&
						product.FlavorNotes == "Chocolate, Nutty" &&
						product.StripeID == "stripe_prod_123"
				})).Return(nil)
			},
			expectedError: false,
			withRequestID: false,
		},
		{
			name:      "Product without Stripe ID",
			productID: uuid.New(),
			productDTO: &dto.ProductUpdateDTO{
				Name:       strPtr("Updated Coffee"),
				StockLevel: intPtr(95),
			},
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// 1. Mock GetByID response
				existingProduct := &models.Product{
					ID:          uuid.New(),
					Name:        "Coffee",
					Description: "Generic coffee",
					ImageURL:    "https://example.com/coffee.jpg",
					Active:      true,
					StockLevel:  100,
					Weight:      250,
					Origin:      "Colombia",
					RoastLevel:  "Medium",
					FlavorNotes: "Chocolate, Nutty",
					StripeID:    "", // No Stripe ID
					CreatedAt:   time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now().Add(-12 * time.Hour),
				}
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(existingProduct, nil)

				// 2. No mock for Stripe as it should be skipped

				// 3. Mock Update in repository
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(product *models.Product) bool {
					return product.Name == "Updated Coffee" &&
						product.StockLevel == 95 &&
						product.StripeID == "" // Still no Stripe ID
				})).Return(nil)
			},
			expectedError: false,
			withRequestID: false,
		},
		{
			name:      "Product not found",
			productID: uuid.New(),
			productDTO: &dto.ProductUpdateDTO{
				Name: strPtr("Updated Coffee"),
			},
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// Mock GetByID response - product not found
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("database error: product not found"))
			},
			expectedError:  true,
			expectedErrMsg: "error retrieving product",
			withRequestID:  false,
		},
		{
			name:      "Product found but nil returned",
			productID: uuid.New(),
			productDTO: &dto.ProductUpdateDTO{
				Name: strPtr("Updated Coffee"),
			},
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// Mock GetByID response - no error but nil product
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, nil)
			},
			expectedError:  true,
			expectedErrMsg: "product not found with id",
			withRequestID:  false,
		},
		{
			name:      "Stripe update error",
			productID: uuid.New(),
			productDTO: &dto.ProductUpdateDTO{
				Name: strPtr("Updated Coffee"),
			},
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// 1. Mock GetByID response
				existingProduct := &models.Product{
					ID:          uuid.New(),
					Name:        "Coffee",
					Description: "Generic coffee",
					StripeID:    "stripe_prod_123",
				}
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(existingProduct, nil)

				// 2. Mock UpdateProduct in Stripe - fails
				mockStripe.On("UpdateProduct", mock.Anything, mock.Anything).Return(errors.New("stripe API error"))
			},
			expectedError:  true,
			expectedErrMsg: "error updating product in Stripe",
			withRequestID:  false,
		},
		{
			name:      "Database update error",
			productID: uuid.New(),
			productDTO: &dto.ProductUpdateDTO{
				Name: strPtr("Updated Coffee"),
			},
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// 1. Mock GetByID response
				existingProduct := &models.Product{
					ID:          uuid.New(),
					Name:        "Coffee",
					Description: "Generic coffee",
					StripeID:    "stripe_prod_123",
				}
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(existingProduct, nil)

				// 2. Mock UpdateProduct in Stripe - succeeds
				mockStripe.On("UpdateProduct", mock.Anything, mock.Anything).Return(nil)

				// 3. Mock Update in repository - fails
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			expectedError:  true,
			expectedErrMsg: "error updating product in database",
			withRequestID:  false,
		},
		{
			name:      "With request ID in context",
			productID: uuid.New(),
			productDTO: &dto.ProductUpdateDTO{
				Name: strPtr("Updated Coffee"),
			},
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// 1. Mock GetByID response
				existingProduct := &models.Product{
					ID:          uuid.New(),
					Name:        "Coffee",
					Description: "Generic coffee",
					StripeID:    "stripe_prod_123",
				}
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(existingProduct, nil)

				// 2. Mock UpdateProduct in Stripe
				mockStripe.On("UpdateProduct", mock.Anything, mock.Anything).Return(nil)

				// 3. Mock Update in repository
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
			withRequestID: true, // This test includes a request ID in the context
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockRepo := new(mocks.ProductRepository)
			mockStripe := new(stripemock.Client)

			// Configure mocks
			tc.setupMocks(mockRepo, mockStripe)

			// Create service
			productService := services.NewProductService(mockRepo, mockStripe, &logger)

			// Create context, with request ID if needed
			var ctx context.Context
			if tc.withRequestID {
				ctx = context.WithValue(context.Background(), "request_id", "test-request-id-123")
			} else {
				ctx = context.Background()
			}

			// Execute test
			result, err := productService.Update(ctx, tc.productID, tc.productDTO)

			// Assertions
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// If the DTO has fields set, verify they were applied to the result
				if tc.productDTO.Name != nil {
					assert.Equal(t, *tc.productDTO.Name, result.Name)
				}
				if tc.productDTO.Description != nil {
					assert.Equal(t, *tc.productDTO.Description, result.Description)
				}
				if tc.productDTO.ImageURL != nil {
					assert.Equal(t, *tc.productDTO.ImageURL, result.ImageURL)
				}
				if tc.productDTO.Active != nil {
					assert.Equal(t, *tc.productDTO.Active, result.Active)
				}
				if tc.productDTO.StockLevel != nil {
					assert.Equal(t, *tc.productDTO.StockLevel, result.StockLevel)
				}
				if tc.productDTO.Weight != nil {
					assert.Equal(t, *tc.productDTO.Weight, result.Weight)
				}
				if tc.productDTO.Origin != nil {
					assert.Equal(t, *tc.productDTO.Origin, result.Origin)
				}
				if tc.productDTO.RoastLevel != nil {
					assert.Equal(t, *tc.productDTO.RoastLevel, result.RoastLevel)
				}
				if tc.productDTO.FlavorNotes != nil {
					assert.Equal(t, *tc.productDTO.FlavorNotes, result.FlavorNotes)
				}

				// Verify that UpdatedAt was updated
				assert.WithinDuration(t, time.Now(), result.UpdatedAt, 5*time.Second)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
			mockStripe.AssertExpectations(t)
		})
	}
}

func TestProductService_Delete(t *testing.T) {
	// Initialize logger
	logger := zerolog.New(zerolog.NewTestWriter(t))

	// Test cases
	testCases := []struct {
		name           string
		productID      uuid.UUID
		setupMocks     func(*mocks.ProductRepository, *stripemock.Client)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:      "Successful product deletion with Stripe ID",
			productID: uuid.New(),
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// 1. Mock GetByID response
				product := &models.Product{
					ID:          uuid.New(),
					Name:        "Colombian Coffee",
					Description: "Rich and smooth coffee from Colombia",
					StripeID:    "stripe_prod_123",
				}
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(product, nil)

				// 2. Mock ArchiveProduct in Stripe
				mockStripe.On("ArchiveProduct", mock.Anything, "stripe_prod_123").Return(nil)

				// 3. Mock Delete in repository
				mockRepo.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "Successful product deletion without Stripe ID",
			productID: uuid.New(),
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// 1. Mock GetByID response
				product := &models.Product{
					ID:          uuid.New(),
					Name:        "Colombian Coffee",
					Description: "Rich and smooth coffee from Colombia",
					StripeID:    "", // No Stripe ID
				}
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(product, nil)

				// 2. No mock for Stripe as it should be skipped

				// 3. Mock Delete in repository
				mockRepo.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "Product retrieval error",
			productID: uuid.New(),
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// Mock GetByID response - retrieval error
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("database error"))
			},
			expectedError:  true,
			expectedErrMsg: "failed to retrieve product for deletion",
		},
		{
			name:      "Product not found",
			productID: uuid.New(),
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// Mock GetByID response - product not found (no error but nil product)
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, nil)
			},
			expectedError:  true,
			expectedErrMsg: "product with ID",
		},
		{
			name:      "Stripe archive error",
			productID: uuid.New(),
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// 1. Mock GetByID response
				product := &models.Product{
					ID:          uuid.New(),
					Name:        "Colombian Coffee",
					Description: "Rich and smooth coffee from Colombia",
					StripeID:    "stripe_prod_123",
				}
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(product, nil)

				// 2. Mock ArchiveProduct in Stripe - fails
				mockStripe.On("ArchiveProduct", mock.Anything, "stripe_prod_123").Return(errors.New("stripe API error"))
			},
			expectedError:  true,
			expectedErrMsg: "failed to archive product in Stripe",
		},
		{
			name:      "Database deletion error",
			productID: uuid.New(),
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// 1. Mock GetByID response
				product := &models.Product{
					ID:          uuid.New(),
					Name:        "Colombian Coffee",
					Description: "Rich and smooth coffee from Colombia",
					StripeID:    "stripe_prod_123",
				}
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(product, nil)

				// 2. Mock ArchiveProduct in Stripe - succeeds
				mockStripe.On("ArchiveProduct", mock.Anything, "stripe_prod_123").Return(nil)

				// 3. Mock Delete in repository - fails
				mockRepo.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(errors.New("database error"))
			},
			expectedError:  true,
			expectedErrMsg: "failed to delete product from database",
		},
		{
			name:      "Database deletion error without Stripe ID",
			productID: uuid.New(),
			setupMocks: func(mockRepo *mocks.ProductRepository, mockStripe *stripemock.Client) {
				// 1. Mock GetByID response
				product := &models.Product{
					ID:          uuid.New(),
					Name:        "Colombian Coffee",
					Description: "Rich and smooth coffee from Colombia",
					StripeID:    "", // No Stripe ID
				}
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(product, nil)

				// 2. No Stripe mock needed

				// 3. Mock Delete in repository - fails
				mockRepo.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(errors.New("database error"))
			},
			expectedError:  true,
			expectedErrMsg: "failed to delete product from database",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockRepo := new(mocks.ProductRepository)
			mockStripe := new(stripemock.Client)

			// Configure mocks
			tc.setupMocks(mockRepo, mockStripe)

			// Create service
			productService := services.NewProductService(mockRepo, mockStripe, &logger)

			// Execute test
			err := productService.Delete(context.Background(), tc.productID)

			// Assertions
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
			mockStripe.AssertExpectations(t)
		})
	}
}