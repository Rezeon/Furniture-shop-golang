package controller

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go-be/database"
	"go-be/models"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// --- Konfigurasi Duitku (GANTI DENGAN NILAI ASLI ANDA!) ---

var DUITKU_MERCHANT_CODE = os.Getenv("DUITKU_MERCHANT_CODE") // Ganti dengan Kode Merchant Anda
var DUITKU_API_KEY = os.Getenv("DUITKU_API_KEY")             // Ganti dengan API Key Anda
var DUITKU_CALLBACK_URL = os.Getenv("DUITKU_CALLBACK_URL")
var DUITKU_RETURN_URL = os.Getenv("DUITKU_RETURN_URL")
var DUITKU_ENDPOINT = os.Getenv("DUITKU_ENDPOINT")

// --- Struktur DTO untuk API Duitku ---

type ItemDetail struct {
	Name     string `json:"name"`
	Price    uint   `json:"price"` // Harga per unit
	Quantity uint   `json:"quantity"`
}

type CreateInvoiceRequest struct {
	PaymentAmount   uint         `json:"paymentAmount"`
	MerchantOrderID string       `json:"merchantOrderId"`
	ProductDetails  string       `json:"productDetails"`
	Email           string       `json:"email"`
	PhoneNumber     string       `json:"phoneNumber"`
	CustomerVaName  string       `json:"customerVaName"`
	ItemDetails     []ItemDetail `json:"itemDetails"`
	CallbackUrl     string       `json:"callbackUrl"`
	ReturnUrl       string       `json:"returnUrl"`
	ExpiryPeriod    int          `json:"expiryPeriod"`
}

type CreateInvoiceResponse struct {
	MerchantCode  string `json:"merchantCode"`
	Reference     string `json:"reference"`
	PaymentUrl    string `json:"paymentUrl"`
	StatusCode    string `json:"statusCode"` // "00" jika sukses
	StatusMessage string `json:"statusMessage"`
}

// DuitkuCallbackInput merepresentasikan parameter yang diterima dari Duitku (x-www-form-urlencoded).
type DuitkuCallbackInput struct {
	MerchantCode    string `form:"merchantCode"`
	Amount          uint   `form:"amount"`
	MerchantOrderID string `form:"merchantOrderId"`
	ProductDetail   string `form:"productDetail"`
	ResultCode      string `form:"resultCode"` // 00: Success, 01: Failed
	Reference       string `form:"reference"`
	Signature       string `form:"signature"`
}

// --- Helper Functions ---

// createDuitkuInvoice mengirim request ke API Duitku untuk membuat invoice.
func createDuitkuInvoice(order models.Order, cart models.Cart, customerEmail string, customerPhone string) (CreateInvoiceResponse, error) {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)

	// 1. Buat Signature SHA256(merchantCode + timestamp + apiKey)
	signatureString := DUITKU_MERCHANT_CODE + timestamp + DUITKU_API_KEY
	hasher := sha256.New()
	hasher.Write([]byte(signatureString))
	signature := hex.EncodeToString(hasher.Sum(nil))

	// 2. Siapkan Payload
	var items []ItemDetail
	var checkTotal uint = 0

	for _, item := range cart.Items {
		itemPrice := item.Product.Price
		itemQuantity := item.Quantity

		checkTotal += itemPrice * itemQuantity

		items = append(items, ItemDetail{
			Name:     item.Product.Name,
			Price:    itemPrice,
			Quantity: itemQuantity,
		})
	}

	if checkTotal != order.TotalPrice {
		return CreateInvoiceResponse{}, fmt.Errorf("Internal data inconsistency: Calculated total price (%d) does not match Order's TotalPrice (%d). Cannot send to Duitku.", checkTotal, order.TotalPrice)
	}

	payload := CreateInvoiceRequest{
		PaymentAmount:   checkTotal,
		MerchantOrderID: strconv.FormatUint(uint64(order.ID), 10),
		ProductDetails:  "Pembayaran Pesanan #" + strconv.FormatUint(uint64(order.ID), 10),
		Email:           customerEmail,
		PhoneNumber:     customerPhone,
		CustomerVaName:  "Customer " + strconv.FormatUint(uint64(order.UserID), 10),
		ItemDetails:     items,
		CallbackUrl:     DUITKU_CALLBACK_URL,
		ReturnUrl:       DUITKU_RETURN_URL,
		ExpiryPeriod:    30, // 30 menit
	}

	payloadBytes, _ := json.Marshal(payload)

	// 3. Kirim HTTP Request
	req, err := http.NewRequest("POST", DUITKU_ENDPOINT, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return CreateInvoiceResponse{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-duitku-signature", signature)
	req.Header.Set("x-duitku-timestamp", timestamp)
	req.Header.Set("x-duitku-merchantcode", DUITKU_MERCHANT_CODE)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return CreateInvoiceResponse{}, fmt.Errorf("Gagal koneksi ke Duitku: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return CreateInvoiceResponse{}, fmt.Errorf("Gagal membaca body respons Duitku: %w", readErr)
	}

	if resp.StatusCode != http.StatusOK {

		var duitkuErrorResp CreateInvoiceResponse
		json.Unmarshal(bodyBytes, &duitkuErrorResp)
		if duitkuErrorResp.StatusMessage != "" {
			return CreateInvoiceResponse{}, errors.New("Duitku API HTTP Status " + strconv.Itoa(resp.StatusCode) + ": " + duitkuErrorResp.StatusMessage)
		}
		return CreateInvoiceResponse{}, fmt.Errorf("Duitku API returned HTTP Status %d. Raw Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var duitkuResponse CreateInvoiceResponse

	if err := json.Unmarshal(bodyBytes, &duitkuResponse); err != nil {
		return CreateInvoiceResponse{}, fmt.Errorf("Gagal mendekode JSON respons Duitku: %w", err)
	}

	if duitkuResponse.StatusCode != "00" {
		return duitkuResponse, errors.New("Duitku API Error (Code: " + duitkuResponse.StatusCode + "): " + duitkuResponse.StatusMessage)
	}

	return duitkuResponse, nil
}

// --- Controller Handlers ---

// HandleDuitkuCallback menerima notifikasi status pembayaran dari Duitku.
// Route: POST /api/v1/duitku/callback
func HandleDuitkuCallback(c *gin.Context) {
	var input DuitkuCallbackInput

	// Binding data dari x-www-form-urlencoded
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid callback data"})
		return
	}

	// 1. Verifikasi Signature Callback (MD5)
	// Formula: MD5(merchantcode + amount + merchantOrderId + merchantKey)
	signatureString := fmt.Sprintf("%s%d%s%s", input.MerchantCode, input.Amount, input.MerchantOrderID, DUITKU_API_KEY)
	hasher := md5.New()
	hasher.Write([]byte(signatureString))
	expectedSignatureHex := hex.EncodeToString(hasher.Sum(nil))

	if input.Signature != expectedSignatureHex {
		//  Jangan kirim status 200 OK jika signature gagal!
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// 2. Cari Order di database berdasarkan MerchantOrderID
	orderID, _ := strconv.ParseUint(input.MerchantOrderID, 10, 64)
	var order models.Order
	if err := database.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// 3. Verifikasi Jumlah (Amount)
	// Pastikan jumlah yang dibayarkan sama dengan TotalPrice di Order Anda
	if input.Amount != order.TotalPrice {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount mismatch"})
		return
	}

	// 4. Proses Status Pembayaran
	newStatus := order.Status

	if input.ResultCode == "00" {
		newStatus = "Paid"
	} else if input.ResultCode == "01" {
		newStatus = "Failed"
	} else {
		// Status lainnya, misalnya 02: Pending
		newStatus = "Pending"
	}

	// Hanya update jika status berubah
	if newStatus != order.Status {
		if err := database.DB.Model(&order).Update("Status", newStatus).Error; err != nil {
			// Jika gagal update DB, Anda bisa log error dan mengembalikan non-200 OK
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
			return
		}
	}

	// 5. Beri Respon HTTP 200 OK ke Duitku
	// Ini memberitahu Duitku bahwa Anda telah menerima dan memproses notifikasi
	c.String(http.StatusOK, "Callback received and processed")
}
