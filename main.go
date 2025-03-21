package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func main() {
	var err error
	DB, err = sql.Open("mysql", "user:1234@tcp(localhost:3306)/breakfast?charset=utf8mb4&parseTime=True")
	if err != nil {
		log.Fatal("無法連接資料庫:", err)
	}
	defer DB.Close()

	if err := DB.Ping(); err != nil {
		log.Fatal("資料庫 ping 失敗:", err)
	}

	router := gin.Default()
	router.Use(cors.Default())

	// 設定首頁（當用戶訪問 "/" 時，回傳 index.html）
	router.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})
	// 提供靜態檔案（HTML, CSS, JS）
	router.Static("/static", "./static")

	router.POST("/add-to-cart", AddToCart)
	router.GET("/get-cart", GetCart)
	router.DELETE("/clear-cart", ClearCart)
	router.DELETE("/remove-from-cart", RemoveFromCart)
	router.PUT("/update-cart", UpdateCart)

	router.Run(":8080")
}

type AddToCartRequest struct {
	ProductID   int    `json:"product_id"`
	TableNumber string `json:"table_number"`
}

func AddToCart(c *gin.Context) {
	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "格式錯誤"})
		return
	}

	// 查詢該商品的單價
	var price int
	err := DB.QueryRow("SELECT price FROM products WHERE product_id = ?", req.ProductID).Scan(&price)
	if err != nil {
		c.JSON(500, gin.H{"error": "商品查詢失敗"})
		return
	}

	_, err = DB.Exec(`
        INSERT INTO cart (product_id, product_name, quantity, total_price, table_number)
        SELECT ?, product_name, 1, price, ?
        FROM products
        WHERE product_id = ?
        ON DUPLICATE KEY UPDATE 
            quantity = quantity + 1,
            total_price = total_price + ?  -- 更新總價 = 舊的總價 + 新增的單價
    `, req.ProductID, req.TableNumber, req.ProductID, price)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "加入購物車成功"})
}

func GetCart(c *gin.Context) {
	tableNumber := c.Query("table_number")
	if tableNumber == "" {
		c.JSON(400, gin.H{"error": "缺少桌號"})
		return
	}

	rows, err := DB.Query(`
        SELECT product_name, quantity, total_price 
        FROM cart 
        WHERE table_number = ?`, tableNumber)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var cart []string
	var totalPrice int

	for rows.Next() {
		var productName string
		var quantity, total int
		err := rows.Scan(&productName, &quantity, &total)
		if err != nil {
			c.JSON(500, gin.H{"error": "讀取購物車失敗"})
			return
		}

		totalPrice += total // 🔥 直接加總 total_price

		cart = append(cart, fmt.Sprintf("%s 數量%d 總計為%d元", productName, quantity, total))
	}

	c.JSON(200, gin.H{
		"cart":        cart,
		"total_price": totalPrice,
	})
}

func ClearCart(c *gin.Context) {
	tableNumber := c.Query("table_number")
	if tableNumber == "" {
		c.JSON(400, gin.H{"error": "缺少桌號"})
		return
	}

	_, err := DB.Exec("DELETE FROM cart WHERE table_number = ?", tableNumber)
	if err != nil {
		c.JSON(500, gin.H{"error": "清空購物車失敗: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "購物車已清空"})
}

// 當我的購物車例如漢堡數量變成0時候，將漢堡從購物車中刪除
func RemoveFromCart(c *gin.Context) {
	productID := c.Query("product_id")
	tableNumber := c.Query("table_number")

	_, err := DB.Exec("DELETE FROM cart WHERE product_id = ? AND table_number = ?", productID, tableNumber)
	if err != nil {
		c.JSON(500, gin.H{"error": "刪除商品失敗"})
		return
	}

	c.JSON(200, gin.H{"message": "商品已移除"})
}

func UpdateCart(c *gin.Context) {
	var req struct {
		ProductID   int    `json:"product_id"`
		TableNumber string `json:"table_number"`
		Quantity    int    `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("JSON 綁定錯誤:", err)
		c.JSON(400, gin.H{"error": "格式錯誤"})
		return
	}

	// 先查詢該商品的單價（用 total_price / quantity 計算）
	var unitPrice float64
	err := DB.QueryRow(`
        SELECT total_price / quantity 
        FROM cart 
        WHERE product_id = ? AND table_number = ?`,
		req.ProductID, req.TableNumber).Scan(&unitPrice)
	if err != nil {
		log.Println("單價查詢錯誤:", err)
		c.JSON(500, gin.H{"error": "無法獲取單價"})
		return
	}

	// 更新購物車數量與總價
	_, err = DB.Exec(`
        UPDATE cart 
        SET quantity = ?, total_price = ? * ?
        WHERE product_id = ? AND table_number = ?`,
		req.Quantity, unitPrice, req.Quantity, req.ProductID, req.TableNumber)

	if err != nil {
		log.Println("SQL 更新錯誤:", err)
		c.JSON(500, gin.H{"error": "更新購物車失敗"})
		return
	}

	c.JSON(200, gin.H{"message": "購物車數量已更新"})
}
