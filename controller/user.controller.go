package controller

import (
	"go-be/database"
	"go-be/models"
	"go-be/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func SignUp(c *gin.Context) {
	var user models.User
	godotenv.Load()

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "please input correctly",
		})
		return
	}

	var existUser models.User

	if err := database.DB.Where("email = ?", user.Email).Find(&existUser).Error; err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"error": "email already taken",
		})
		return
	}

	hashPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to bcrypt",
		})
		return
	}

	user.Password = string(hashPassword)
	user.Role = string("user")

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server failed to create user"})
		return
	}
	tokenString, err := utils.GenerateToken(user.ID)
	//secret := os.Getenv(("JWT_SECRET"))

	//claim := jwt.MapClaims{
	//  "user_id": user.ID,
	//  "exp":     time.Now().Add(time.Hour * 24).Unix(),
	//}
	//token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	//tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"token":   tokenString,
	})

}

func SignIn(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "please input correctly",
		})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	err := utils.ReverseHash(input.Password, user.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}

	tokenString, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User login successfully",
		"token":   tokenString,
	})
}
func GetUserById(c *gin.Context) {
	var user models.User

	userId, exists := c.Get("userId")

	if !exists {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "must be user or admin to get informatin"})
		return
	}

	if err := database.DB.Preload("Address").First(&user, userId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found in database"})
		return
	}

	c.JSON(http.StatusOK, user)
}
func GetUser(c *gin.Context) {
	var user []models.User

	if err := database.DB.Find(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error server"})
	}

	c.JSON(http.StatusOK, user)
}
func UpdateUser(c *gin.Context) {
	userID, existUser := c.Get("userId")
	if !existUser {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Name      string `json:"name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		AddressId *uint  `json:"addressId"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.AddressId != nil {
		user.AddressID = input.AddressId
	}
	if input.Password != "" {
		hash, err := utils.HashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		user.Password = hash
	}

	// Simpan ke database
	if err := database.DB.Model(&user).Updates(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user updated successfully",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func DeleteUser(c *gin.Context) {
	id, existUser := c.Get("user_id")
	if !existUser {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
