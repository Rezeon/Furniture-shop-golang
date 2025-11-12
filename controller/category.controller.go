package controller

import (
	"go-be/database"
	"go-be/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := database.DB.Where("name = ?", category.Name).FirstOrCreate(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "category created or already exists",
		"data":    category,
	})

}
func GetCategory(c *gin.Context) {
	var category []models.Category

	if err := database.DB.Find(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, category)
}
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	if err := database.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category tidak ditemukan"})
		return
	}

	database.DB.Delete(&category)
}
