package main

import (
    "net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"strings"
)

// User represents a user in our system
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code,omitempty"`
}

// In-memory storage
var users = []User{
	{ID: 1, Name: "John Doe", Email: "john@example.com", Age: 30},
	{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Age: 25},
	{ID: 3, Name: "Bob Wilson", Email: "bob@example.com", Age: 35},
}
var nextID = 4

func main(){
	// TODO: Create Gin router
	router := gin.Default()

	// TODO: Setup routes
	router.GET("/users",getAllUsers) // GET /users - Get all users
	router.GET("/users/:id",getUserByID) // GET /users/:id - Get user by ID
	router.POST("/users",createUser) // POST /users - Create new user
	router.PUT("/users/:id",updateUser) // PUT /users/:id - Update user
	router.DELETE("/users/:id",deleteUser) // DELETE /users/:id - Delete user
	router.GET("/users/search",searchUsers)  // GET /users/search - Search users by name

	// TODO: Start server on port 8080
	router.Run(":8080")
}
// TODO: Implement handler functions

// getAllUsers handles GET /users
func getAllUsers(c *gin.Context) {
	// TODO: Return all users
	c.JSON(http.StatusOK, users)
	
	
}

// getUserByID handles GET /users/:id
func getUserByID(c *gin.Context) {
    idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Error:"Invalid user ID"})
		return
	}

	// 2. Iterate through the slice to find the user
	for _, user := range users {
		if user.ID == id {
			c.JSON(http.StatusOK, user)
			return
		}
	}
	// 3. Return 404 if not found
	c.JSON(http.StatusNotFound, Response{Error:"User not found"})
}

// createUser handles POST /users
func createUser(c *gin.Context) {
	var newUser User

	// Bind request body to newUser struct
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, Response{Error:"Unexpected error!"})
		return
	}

	// Add new user to the local slice
	newUser.ID = nextID
	users = append(users, newUser)
	nextID++

	// Return the newly created user
	c.JSON(http.StatusCreated, newUser)
}

// updateUser handles PUT /users/:id
func updateUser(c *gin.Context) {
    idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Error:"Invalid User ID"})
		return
	}

	var updatedUser User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, Response{Error:"Unexpected Error!"})
		return
	}

	// Iterate through the slice to find the user [3]
	found := false
	for i, user := range users {
		if user.ID == userID {
			// Update the user in the slice
			users[i].Name = updatedUser.Name
			users[i].Email = updatedUser.Email
			users[i].Age = updatedUser.Age
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, Response{Error:"User not found"})
		return
	}

	c.JSON(http.StatusOK, Response{Message:"User updated", Data:updatedUser})

}

// deleteUser handles DELETE /users/:id
func deleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Error:"Invalid user ID"})
		return
	}

	// Find and remove the user
	for i, user := range users {
		if user.ID == id {
			// Remove the user from the slice
			users = append(users[:i], users[i+1:]...)
			c.JSON(http.StatusOK, Response{Message:"User deleted successfully"})
			return
		}
	}

	c.JSON(http.StatusNotFound, Response{Error:"User not found"})
}

// searchUsers handles GET /users/search?name=value
func searchUsers(c *gin.Context) {
// 1. Get search parameter from query (e.g., /user?id=2)
	nameQuery := c.Query("name")
    var foundUsers []User
    for _, user := range users {
    // Use strings.Contains for a partial match, and strings.ToLower for case-insensitivity
        if strings.Contains(strings.ToLower(user.Name), strings.ToLower(nameQuery)) {
            foundUsers = append(foundUsers, user)
        }
    }
    c.JSON(http.StatusOK, foundUsers)
}