package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"go-be/database"
	"go-be/models"
	"go-be/utils"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var ctx = context.Background()

func CreateProduct(c *gin.Context) {

	var input struct {
		Name        string `form:"name" json:"name"`
		Price       uint   `form:"price" json:"price"`
		Description string `form:"description" json:"description"`
		CategoryID  uint   `form:"categoryId" json:"categoryId"`
	}
	// masukan data ke input

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusNoContent, gin.H{"error": "please input correctly"})
		return
	}
	// masukan data input tadi ke moodel database
	product := models.Product{
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
		CategoryID:  input.CategoryID,
	}

	file, err := c.FormFile("image")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please input image"})
		return
	}
	// buat file sementara
	tempPath := "./tempImage/" + file.Filename
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	//upload image ke claudinary
	url, publicID, err := utils.UploadImage(tempPath, "image")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	product.Image = url
	product.PublicID = publicID
	//delete file sementara tadi
	os.Remove(tempPath)

	if err := database.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "product create",
		"data":    product,
	})
}

func GetProduct(c *gin.Context) {
	var product []models.Product
	chaceKey := "product:list:10"
	chacheList, err := utils.RedisClient.Get(ctx, chaceKey).Result()
	if err == nil {
		log.Printf("INFO: Cache Hit untuk kunci: %s", chaceKey)
		c.JSON(http.StatusOK, json.RawMessage(chacheList))
		return
	}
	if err := database.DB.Preload("Category").Find(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	productJson, _ := json.Marshal(product)
	utils.RedisClient.Set(ctx, chaceKey, productJson, 5*time.Minute)

	c.JSON(http.StatusOK, product)
}

func GetProductByID(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	cacheKey := fmt.Sprintf("product:%s", id)

	chacheData, err := utils.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		log.Printf("INFO: Cache Hit untuk kunci: %s", chacheData)
		c.JSON(http.StatusOK, chacheData)
		return
	}
	if err := database.DB.Preload("Category").First(&product, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server internal error"})
		return
	}
	productJson, _ := json.Marshal(product)
	utils.RedisClient.Set(ctx, cacheKey, productJson, 5*time.Minute)
	c.JSON(http.StatusOK, product)
}

func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	cacheKey := fmt.Sprintf("product:%s", id)
	//  Cari produk berdasarkan ID
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	//  Ambil form field
	name := c.PostForm("name")
	price := c.PostForm("price")
	description := c.PostForm("description")
	category := c.PostForm("categoryId")

	//  Ambil file upload
	file, _ := c.FormFile("image")

	product.Name = name
	product.Price = utils.StringToUint(price)
	product.Description = description
	product.CategoryID = utils.StringToUint(category)
	fmt.Printf("Received Form Data: Name=%s, Price=%s, Description=%s, CategoryID=%s, ImageFileExists=%t\n",
		name, price, description, category, file != nil)
	//  Jika ada file image baru
	if file != nil {
		tempPath := "./tempImage/" + file.Filename
		// hapus file yang ada di database
		utils.DeleteImage(product.PublicID)
		// Simpan file sementara
		if err := c.SaveUploadedFile(file, tempPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
			return
		}

		// Upload ke Cloudinary atau storage lain
		url, publicID, err := utils.UploadImage(tempPath, "products")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		product.Image = url
		product.PublicID = publicID
		// Hapus file sementara
		os.Remove(tempPath)
	}

	// Update data ke database
	database.DB.Save(&product)
	database.DB.First(&product, id)
	_, redisErr := utils.RedisClient.Del(ctx, cacheKey).Result()
	if redisErr != nil {
		log.Printf("Warning: Gagal menghapus cache saat update: %v", redisErr)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "product updated successfully",
		"data":    product,
	})
}

func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	cacheKey := fmt.Sprintf("product:%s", id)
	var product models.Product
	// cari product berdasarkan id
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error for find data"})
		return
	}
	//hapus image dari claudinary
	utils.DeleteImage(product.PublicID)
	//delete product dari database
	_, redisErr := utils.RedisClient.Del(ctx, cacheKey).Result()
	if redisErr != nil {
		log.Printf("Warning: Gagal menghapus cache saat update: %v", redisErr)
	}
	database.DB.Delete(&product)
	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}
