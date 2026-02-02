package main
import (
	"regexp"
	"strings"
	"time"
	
)
var products = []Product{
	{
		ID:    1,
		SKU:   "ELE-101-WHT",
		Name:  "Wireless Headphones",
		Price: 99.99,
		Currency: "USD",
		Category: Category{ID: 1, Name: "Electronics", Slug: "electronics"},
		Inventory: Inventory{
			Quantity: 50,
			Reserved: 5,
			Location: "WH001",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		ID:    2,
		SKU:   "HOM-202-BLU",
		Name:  "Smart Coffee Maker",
		Price: 149.50,
		Currency: "USD",
		Category: Category{ID: 4, Name: "Home & Garden", Slug: "home-garden"},
		Inventory: Inventory{
			Quantity: 20,
			Reserved: 2,
			Location: "WH002",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
}

// Update the next ID to avoid conflicts
var nextProductID = 3
func isValidSKU(sku string) bool {
	re := regexp.MustCompile(`^[A-Z]{3}-\d{3}-[A-Z]{3}$`)
	return re.MatchString(sku)
}

func isValidCurrency(currency string) bool {
	for _, v := range validCurrencies {
		if v == strings.ToUpper(currency) {
			return true
		}
	}
	return false
}

func isValidCategory(categoryName string) bool {
	for _, cat := range categories {
		if strings.EqualFold(cat.Name, categoryName) {
			return true
		}
	}
	return false
}

func isValidSlug(slug string) bool {
	re := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	return re.MatchString(slug)
}

func isValidWarehouseCode(code string) bool {
	for _, v := range validWarehouses {
		if v == code {
			return true
		}
	}
	return false
}
func validateProduct(product *Product) []ValidationError {
	var errors []ValidationError

	if !isValidSKU(product.SKU) {
		errors = append(errors, ValidationError{Field: "sku", Tag: "format", Message: "Invalid SKU pattern"})
	}
	
	// Uniqueness check
	for _, p := range products {
		if p.SKU == product.SKU && p.ID != product.ID {
			errors = append(errors, ValidationError{Field: "sku", Tag: "unique", Message: "SKU already exists"})
		}
	}

	if !isValidCurrency(product.Currency) {
		errors = append(errors, ValidationError{Field: "currency", Tag: "enum", Message: "Unsupported currency"})
	}

	if product.Inventory.Reserved > product.Inventory.Quantity {
		errors = append(errors, ValidationError{Field: "inventory.reserved", Tag: "lte", Message: "Reserved cannot exceed quantity"})
	}

	return errors
}

func sanitizeProduct(product *Product) {
	product.SKU = strings.TrimSpace(strings.ToUpper(product.SKU))
	product.Name = strings.TrimSpace(product.Name)
	product.Currency = strings.ToUpper(product.Currency)
	product.Category.Slug = strings.ToLower(product.Category.Slug)
	
	// Calculated fields
	product.Inventory.Available = product.Inventory.Quantity - product.Inventory.Reserved
	
	now := time.Now()
	if product.CreatedAt.IsZero() {
		product.CreatedAt = now
	}
	product.UpdatedAt = now
	product.Inventory.LastUpdated = now
}
func main() {
	r := gin.New()
	
	// Apply your custom middlewares
	r.Use(ErrorHandlerMiddleware()) 
	r.Use(CORSMiddleware("GET, POST, OPTIONS"))

	// Product Routes
	r.POST("/products", createProduct)
	r.POST("/products/bulk", createProductsBulk)
	r.POST("/categories", createCategory)
	
	r.Run(":8080")
}

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Recovery from panics
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic: %v", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, APIResponse{
					Success: false,
					Error:   "500",
					Message: "Internal Server Error",
				})
			}
		}()

		c.Next()

		// Handle logical errors attached via c.Error()
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			c.JSON(c.Writer.Status(), APIResponse{
				Success: false,
				Message: err.Error(),
				Error:   strconv.Itoa(c.Writer.Status()),
			})
		}
	}
}
func CORSMiddleware(allowedMethods string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Request-ID, Authorization")
		
		// Set the methods based on the input string
		c.Writer.Header().Set("Access-Control-Allow-Methods", allowedMethods)

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}