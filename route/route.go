package route

import (
	"go-be/controller"
	"go-be/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoute() *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.POST("/sign-up", controller.SignUp)
	r.POST("/sign-in", controller.SignIn)
	r.GET("/product", controller.GetProduct)
	r.GET("/product/:id", controller.GetProductByID)
	r.POST("/api/v1/duitku/callback", controller.HandleDuitkuCallback)
	r.GET("/category", controller.GetCategory)

	userRoute := r.Group("/users", middleware.AuthMiddleware())
	{
		userRoute.GET("", controller.GetUserById)
		userRoute.GET("/admin", controller.GetUser)
		userRoute.PUT("/update", controller.UpdateUser)
		userRoute.DELETE("/delete", controller.DeleteUser)
		userRoute.GET("/address", controller.GetAddress)
		userRoute.POST("/create-address", controller.CreateAddress)
		userRoute.PUT("/update-address", controller.UpdateAddress)
		userRoute.DELETE("/delete-address", controller.DeleteAddress)

	}
	productRoute := r.Group("/product-admin", middleware.AuthMiddleware(), middleware.AdminMiddleware)
	{
		productRoute.POST("/create", controller.CreateProduct)
		productRoute.PUT("/update/:id", controller.UpdateProduct)
		productRoute.DELETE("/delete/:id", controller.DeleteProduct)
	}
	categoryRoute := r.Group("/category-admin", middleware.AuthMiddleware(), middleware.AdminMiddleware)
	{
		categoryRoute.POST("/create", controller.CreateCategory)
		categoryRoute.DELETE("/delete/:id", controller.DeleteCategory)
	}
	cartRoute := r.Group("/cart", middleware.AuthMiddleware())
	{
		cartRoute.GET("", controller.GetUserCart)
		cartRoute.POST("/create", controller.AddToCart)
		cartRoute.PUT("/update/:id", controller.UpdateCartItem)
		cartRoute.DELETE("/delete/:id", controller.DeleteCartItem)
		cartRoute.GET("/cart-item/:id", controller.GetCartItemByID)
		cartRoute.PUT("/update-cart/:id", controller.UpdateCartItemQuantity)
		cartRoute.DELETE("/delete-cart/:id", controller.DeleteCartItem)
	}
	orderRoute := r.Group("/oder", middleware.AuthMiddleware())
	{
		orderRoute.POST("/checkout", middleware.LimitByIP(), controller.Checkout)
		orderRoute.GET("/", controller.GetUserOrders)
		orderRoute.GET("/:id", controller.GetOrderByID)
	}

	return r
}
