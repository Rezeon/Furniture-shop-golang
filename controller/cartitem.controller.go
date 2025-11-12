package controller

import (
	"go-be/database"
	"go-be/models"
	"go-be/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UpdateQuantityInput digunakan untuk memperbarui kuantitas item keranjang.
type UpdateQuantityInput struct {
	Quantity uint `json:"quantity" binding:"required,min=1"`
}

// GetCartItemByID mengambil detail spesifik dari satu CartItem berdasarkan ID-nya.
// Route: GET /api/v1/cart-item/:id
func GetCartItemByID(c *gin.Context) {
	Id, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "harus login dulu"})
	}
	userID := utils.InterfaceToUint(Id)

	cartItemIDStr := c.Param("id")
	cartItemID, err := strconv.ParseUint(cartItemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID item keranjang tidak valid"})
		return
	}

	var cartItem models.CartItem

	// Cari CartItem dan pastikan item tersebut ada di keranjang AKTIF (order_id IS NULL) milik pengguna yang benar.
	if err := database.DB.
		Preload("Product"). // Memuat detail produk
		Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("cart_items.id = ? AND carts.user_id = ? AND carts.order_id IS NULL", cartItemID, userID).
		First(&cartItem).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item keranjang tidak ditemukan atau bukan milik Anda"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil detail item keranjang"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Detail item keranjang berhasil diambil",
		"data":    cartItem,
	})
}

// UpdateCartItemQuantity memperbarui jumlah (Quantity) dari CartItem tertentu.
// Route: PUT /api/v1/cart-item/:id
func UpdateCartItemQuantity(c *gin.Context) {
	Id, exists := c.Get("userId") //  Placeholder, Ganti dengan logika autentikasi yang sebenarnya!
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "harus login dulu"})
	}
	userID := utils.InterfaceToUint(Id)
	cartItemIDStr := c.Param("id")
	cartItemID, err := strconv.ParseUint(cartItemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID item keranjang tidak valid"})
		return
	}

	var input UpdateQuantityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input kuantitas tidak valid", "details": err.Error()})
		return
	}

	var cartItem models.CartItem
	// Cari item, pastikan item tersebut ada di keranjang AKTIF (order_id IS NULL) milik pengguna yang benar.
	result := database.DB.
		Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("cart_items.id = ? AND carts.user_id = ? AND carts.order_id IS NULL", cartItemID, userID).
		First(&cartItem)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item keranjang tidak ditemukan atau bukan milik Anda"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencari item keranjang"})
		return
	}

	// Perbarui kuantitas
	if err := database.DB.Model(&cartItem).Update("quantity", input.Quantity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui kuantitas item"})
		return
	}

	// Ambil ulang item setelah update untuk respon
	database.DB.Preload("Product").First(&cartItem, cartItem.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Kuantitas item keranjang berhasil diperbarui", "data": cartItem})
}

// DeleteCartItem menghapus CartItem tertentu.
// Route: DELETE /api/v1/cart-item/:id
func DeleteCartItem(c *gin.Context) {
	Id, exists := c.Get("userId") //  Placeholder, Ganti dengan logika autentikasi yang sebenarnya!
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "harus login dulu"})
	}
	userID := utils.InterfaceToUint(Id)
	cartItemIDStr := c.Param("id")
	cartItemID, err := strconv.ParseUint(cartItemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID item keranjang tidak valid"})
		return
	}

	var cartItem models.CartItem
	// Cari item, pastikan item tersebut ada di keranjang AKTIF (order_id IS NULL) milik pengguna yang benar.
	result := database.DB.
		Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("cart_items.id = ? AND carts.user_id = ? AND carts.order_id IS NULL", cartItemID, userID).
		First(&cartItem)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item keranjang tidak ditemukan atau bukan milik Anda"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencari item keranjang"})
		return
	}

	// Hapus CartItem
	if err := database.DB.Delete(&cartItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus item dari keranjang"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item keranjang berhasil dihapus"})
}
