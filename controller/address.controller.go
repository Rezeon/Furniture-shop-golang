package controller

import (
	"go-be/database"
	"go-be/models"
	"go-be/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateAddress(c *gin.Context) {
	var address models.Address
	userID, existUser := c.Get("userId")
	if !existUser {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	address.UserID = utils.InterfaceToUint(userID)
	if err := c.ShouldBindJSON(&address); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error input addres"})
		return
	}

	if err := database.DB.Create(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error bind data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messsage": "Create address successfuly",
		"data":     address,
	})

}
func GetAddress(c *gin.Context) {
	var address models.Address
	id, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "user id not found"})
		return
	}

	if err := database.DB.Where("userId = ?", id).Find(&address).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
		return
	}
	c.JSON(http.StatusOK, address)
}

func UpdateAddress(c *gin.Context) {
	id := c.Param("id")
	var address models.Address

	if err := database.DB.First(&address, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
	}

	database.DB.Save(&address)

	c.JSON(http.StatusOK, address)
}

func DeleteAddress(c *gin.Context) {
	id := c.Param("id")
	var addres models.Address

	if err := database.DB.First(&addres, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "addres tidak ditemukan"})
		return
	}

	database.DB.Delete(&addres)

	c.JSON(http.StatusOK, gin.H{"message": "success deleted"})
}
