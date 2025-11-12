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

func Checkout(c *gin.Context) {
	Id, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "harus login dulu"})
		return
	}
	userID := utils.InterfaceToUint(Id)

	var user models.User

	if err := database.DB.Preload("Address").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "harus login dulu"})
		return
	}
	var activeCart models.Cart
	// Cari Cart Aktif (order_id IS NULL)
	err := database.DB.
		Preload("Items.Product").
		Where("user_id = ? AND order_id IS NULL", userID).
		First(&activeCart).Error

	if err != nil {
		status := http.StatusInternalServerError
		if err == gorm.ErrRecordNotFound {
			status = http.StatusBadRequest
			c.JSON(status, gin.H{"error": "Keranjang belanja kosong atau tidak ditemukan"})
			return
		}
		c.JSON(status, gin.H{"error": "Gagal mengambil keranjang aktif"})
		return
	}

	if len(activeCart.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keranjang belanja tidak memiliki item"})
		return
	}

	// Hitung Total Harga dan Total Kuantitas
	var totalPrice uint = 0
	var totalQuantity uint = 0
	for _, item := range activeCart.Items {
		totalPrice += item.Product.Price * item.Quantity
		totalQuantity += item.Quantity
	}

	var newOrder models.Order
	// Mulai Transaksi GORM
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		//  Buat Order Baru (Status default harus "Pending")
		newOrder = models.Order{
			UserID:     userID,
			TotalPrice: totalPrice,
			Quantity:   totalQuantity,
			Status:     "Pending", // Set status awal
		}
		if err := tx.Create(&newOrder).Error; err != nil {
			return err
		}

		//  Kaitkan Cart Aktif dengan Order Baru
		if err := tx.Model(&activeCart).
			Select("OrderID").
			Updates(models.Cart{OrderID: &newOrder.ID}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaksi checkout gagal", "details": err.Error()})
		return
	}

	//  Panggil API Duitku Create Invoice
	// Ambil detail user untuk mengisi email dan phone Duitku
	customerEmail := user.Email
	customerPhone := user.Address.PhoneNumber

	duitkuResp, err := createDuitkuInvoice(newOrder, activeCart, customerEmail, customerPhone)

	if err != nil {
		// Jika Duitku gagal, rollback order (Opsional, tapi disarankan)
		database.DB.Delete(&newOrder)
		database.DB.Model(&activeCart).Update("OrderID", nil) // Lepas keterkaitan cart
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat invoice pembayaran", "details": err.Error()})
		return
	}

	//  Simpan DUITKU_REFERENCE ke database Order
	if err := database.DB.Model(&newOrder).
		Update("DuitkuReference", duitkuResp.Reference).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan referensi Duitku"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoice pembayaran berhasil dibuat.",
		"order": gin.H{
			"id":         newOrder.ID,
			"totalPrice": totalPrice,
			"reference":  duitkuResp.Reference,
			"paymentUrl": duitkuResp.PaymentUrl, // URL untuk diarahkan/pop-up
		},
	})
}

// Checkout mengubah Cart aktif pengguna menjadi Order baru.
// Route: POST /api/v1/checkout
//func Checkout(c *gin.Context) {
//	Id, exists := c.Get("user_id")
//	if !exists {
//		c.JSON(http.StatusUnauthorized, gin.H{"error": "harus login dulu"})
//	}
//	userID := utils.InterfaceToUint(Id)
//
//	var input CheckoutInput
//	if err := c.ShouldBindJSON(&input); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid", "details": err.Error()})
//		return
//	}
//
//	var activeCart models.Cart
//	//  Cari Cart Aktif (order_id IS NULL) pengguna dan preload items serta produk
//	err := database.DB.
//		Preload("Items.Product").
//		Where("user_id = ? AND order_id IS NULL", userID).
//		First(&activeCart).Error
//
//	if err == gorm.ErrRecordNotFound {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Keranjang belanja kosong atau tidak ditemukan"})
//		return
//	} else if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil keranjang aktif"})
//		return
//	}
//
//	if len(activeCart.Items) == 0 {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Keranjang belanja tidak memiliki item"})
//		return
//	}
//
//	// Hitung Total Harga dan Total Kuantitas
//	var totalPrice uint = 0
//	var totalQuantity uint = 0
//
//	for _, item := range activeCart.Items {
//		// Product memiliki field Price
//		totalPrice += item.Product.Price * item.Quantity
//		totalQuantity += item.Quantity
//	}
//
//	//  Mulai Transaksi GORM
//	err = database.DB.Transaction(func(tx *gorm.DB) error {
//		//  Buat Order Baru
//		newOrder := models.Order{
//			UserID:     userID,
//			TotalPrice: totalPrice,
//			Payment:    input.Payment,
//			Quantity:   totalQuantity,
//		}
//		if err := tx.Create(&newOrder).Error; err != nil {
//			return err // Rollback jika gagal
//		}
//
//		// Kaitkan Cart Aktif dengan Order Baru
//		if err := tx.Model(&activeCart).
//			Select("OrderID").
//			Updates(models.Cart{OrderID: &newOrder.ID}).Error; err != nil {
//			return err // Rollback jika gagal
//		}
//
//		return nil
//	})
//
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaksi checkout gagal", "details": err.Error()})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"message": "Checkout berhasil! Pesanan telah dibuat.",
//		"order": gin.H{
//			"id":            activeCart.OrderID,
//			"totalPrice":    totalPrice,
//			"totalQuantity": totalQuantity,
//			"payment":       input.Payment,
//		},
//	})
//}

// GetUserOrders mengambil daftar semua pesanan yang dimiliki pengguna.
// Route: GET /api/v1/orders
func GetUserOrders(c *gin.Context) {
	Id, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "harus login dulu"})
	}
	userID := utils.InterfaceToUint(Id)
	var orders []models.Order

	// Cari semua Order milik pengguna, preload Cart dan Cart Items di dalamnya.
	if err := database.DB.
		Preload("Cart.Items.Product").
		Where("user_id = ?", userID).
		Find(&orders).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar pesanan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Daftar pesanan berhasil diambil",
		"data":    orders,
	})
}

// GetOrderByID mengambil detail spesifik dari satu pesanan.
// Route: GET /api/v1/orders/:id
func GetOrderByID(c *gin.Context) {
	Id, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "harus login dulu"})
	}
	userID := utils.InterfaceToUint(Id)
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID pesanan tidak valid"})
		return
	}

	var order models.Order
	// Cari Order berdasarkan ID dan pastikan ia milik pengguna yang benar.
	if err := database.DB.
		Preload("Cart.Items.Product").
		Where("id = ? AND user_id = ?", orderID, userID).
		First(&order).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan atau bukan milik Anda"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil detail pesanan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Detail pesanan berhasil diambil",
		"data":    order,
	})
}
