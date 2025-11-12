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

// Struct Input DTO

// AddToCartInput adalah struktur data yang diterima saat menambah/mengubah item keranjang.
type AddToCartInput struct {
	ProductID uint `json:"productId" binding:"required"`
	Quantity  uint `json:"quantity" binding:"required,min=1"`
}

// UpdateCartItemInput adalah struktur data yang diterima saat memperbarui kuantitas item.
type UpdateCartItemInput struct {
	Quantity uint `json:"quantity" binding:"required,min=1"`
}

// Helper Functions

// getOrCreateUserCart mencari keranjang aktif  pengguna atau membuatnya jika belum ada.
func getOrCreateUserCart(userID uint, db *gorm.DB) (models.Cart, error) {
	var cart models.Cart
	// Cari keranjang yang belum dikaitkan dengan Order (aktif)
	err := db.Where("user_id = ? AND order_id IS NULL", userID).First(&cart).Error

	if err == gorm.ErrRecordNotFound {
		// Jika tidak ditemukan, buat keranjang baru
		cart = models.Cart{UserID: userID}
		if createErr := db.Create(&cart).Error; createErr != nil {
			return models.Cart{}, createErr
		}
		return cart, nil
	} else if err != nil {
		return models.Cart{}, err
	}
	return cart, nil
}

// Controller Handlers

// AddToCart menambahkan produk ke keranjang pengguna.
// Jika produk sudah ada, kuantitas akan ditambahkan.
func AddToCart(c *gin.Context) {
	Id, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "harus login dulu"})
	}
	userID := utils.InterfaceToUint(Id)
	var input AddToCartInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid", "details": err.Error()})
		return
	}

	// Cek produk dengan ProductID benar-benar ada
	var product models.Product
	if err := database.DB.First(&product, input.ProductID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kesalahan server saat memeriksa produk"})
		return
	}

	//  Dapatkan atau buat Keranjang Aktif Pengguna
	cart, err := getOrCreateUserCart(userID, database.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendapatkan/membuat keranjang"})
		return
	}

	//  Cek apakah item sudah ada di keranjang
	var cartItem models.CartItem
	result := database.DB.
		Where("cart_id = ? AND product_id = ?", cart.ID, input.ProductID).
		First(&cartItem)

	if result.Error == gorm.ErrRecordNotFound {
		//  Jika Item tidak ada, buat CartItem baru
		newCartItem := models.CartItem{
			CartID:    cart.ID,
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
		}
		if err := database.DB.Create(&newCartItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menambah item ke keranjang"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Produk berhasil ditambahkan ke keranjang", "data": newCartItem})

	} else if result.Error == nil {
		//  Jika Item ADA, perbarui Quantity (tambahkan kuantitas baru)
		newQuantity := cartItem.Quantity + input.Quantity
		if err := database.DB.Model(&cartItem).Update("quantity", newQuantity).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui kuantitas item"})
			return
		}
		// Ambil ulang item setelah update untuk respon
		database.DB.First(&cartItem, cartItem.ID)
		c.JSON(http.StatusOK, gin.H{"message": "Kuantitas produk berhasil ditambahkan", "data": cartItem})

	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan saat mencari item keranjang"})
		return
	}
}

// GetUserCart mengambil detail keranjang aktif pengguna saat ini, termasuk item dan detail produk.
func GetUserCart(c *gin.Context) {
	Id, exists := c.Get("userId") //  Placeholder, Ganti dengan logika autentikasi yang sebenarnya!
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "harus login dulu"})
	}
	userID := utils.InterfaceToUint(Id)
	var cart models.Cart
	var cartItem []models.CartItem

	// Cari keranjang aktif pengguna dan preload itemnya beserta detail Produk
	if err := database.DB. // Preload CartItem, dan di dalamnya Preload Product
				Where("user_id = ? AND order_id IS NULL", userID).
				First(&cart).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			// Mengembalikan keranjang kosong sebagai respons jika tidak ditemukan
			c.JSON(http.StatusOK, gin.H{"message": "Keranjang belanja kosong", "data": models.Cart{Items: []models.CartItem{}}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil keranjang"})
		return
	}
	if err := database.DB.Preload("Product").Model(&cartItem).Where("cart_id = ? ", cart.ID).Find(&cartItem).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			// Mengembalikan keranjang kosong sebagai respons jika tidak ditemukan
			c.JSON(http.StatusOK, gin.H{"message": "Keranjang belanja kosong", "data": models.Cart{Items: []models.CartItem{}}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil keranjang"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "Detail keranjang berhasil diambil",
		"data":     cart,
		"dataCart": cartItem,
	})
}

// UpdateCartItem memperbarui kuantitas item keranjang tertentu.
func UpdateCartItem(c *gin.Context) {
	Id, exists := c.Get("userId") //  Placeholder, Ganti dengan logika autentikasi yang sebenarnya!
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "harus login dulu"})
	}
	userID := utils.InterfaceToUint(Id)
	// Ambil ID CartItem dari URL parameter
	cartItemIDStr := c.Param("id")
	cartItemID, err := strconv.ParseUint(cartItemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID item keranjang tidak valid"})
		return
	}

	var input UpdateCartItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input kuantitas tidak valid", "details": err.Error()})
		return
	}

	var cartItem models.CartItem
	// Cari CartItem dan pastikan ia milik keranjang AKTIF pengguna yang benar
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

// RemoveCartItem menghapus item dari keranjang berdasarkan ID CartItem.
func RemoveCartItem(c *gin.Context) {
	Id, exists := c.Get("userId") //  Placeholder, Ganti dengan logika autentikasi yang sebenarnya!
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "harus login dulu"})
	}
	userID := utils.InterfaceToUint(Id)
	// Ambil ID CartItem dari URL parameter
	cartItemIDStr := c.Param("id")
	cartItemID, err := strconv.ParseUint(cartItemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID item keranjang tidak valid"})
		return
	}

	var cartItem models.CartItem
	// Cari CartItem dan pastikan ia milik keranjang AKTIF pengguna yang benar
	// Ini memastikan pengguna hanya dapat menghapus item dari keranjang mereka yang aktif.
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
