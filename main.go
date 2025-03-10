package main

import (
	"breakfast-shop/api"
	"breakfast-shop/mysql"

	"github.com/gin-gonic/gin"
)

func main() {
	mysql.InitDB() // 初始化 MySQL 連線

	// 🔹 確保 `mysql.DB` 不是 nil
	if mysql.DB == nil {
		panic(" MySQL 連線失敗，無法啟動伺服器")
	}

	defer mysql.DB.Close() // 讓 MySQL 連線在程式結束時關閉

	r := gin.Default()

	// 提供靜態檔案（HTML, CSS, JS）
	r.Static("/static", "./static")

	// 設定首頁（當用戶訪問 "/" 時，回傳 index.html）
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// 註冊 API 路由
	r.GET("/products", api.GetProducts)

	r.POST("/cart/add", api.AddToCart)
	r.GET("/cart", api.GetCart)
	r.POST("/cart/remove", api.RemoveFromCart)
	r.POST("/cart/clear", api.ClearCart)
	r.POST("/cart/update", api.UpdateCart)
	r.POST("/checkout", api.Checkout)

	// 啟動伺服器
	r.Run(":8080")
}
