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
	//CORS（Cross-Origin Resource Sharing）是瀏覽器的安全機制，預設情況下，前端不能從不同的網域 / 來源去存取後端 API。
	//這裡設定Default的話，我的API對所有的前端都開放，避免出現cors問題訊息
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
	router.POST("/add-batch-to-cart", AddBatchToCart)
	router.POST("/submit-order", SubmitOrder)

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
	var TotalPrice int

	for rows.Next() {
		var productName string
		var quantity, total int
		err := rows.Scan(&productName, &quantity, &total)
		if err != nil {
			c.JSON(500, gin.H{"error": "讀取購物車失敗"})
			return
		}

		TotalPrice += total // 🔥 直接加總 total_price

		cart = append(cart, fmt.Sprintf("%s 數量%d 總計為%d元", productName, quantity, total))
	}

	c.JSON(200, gin.H{
		"cart":        cart,
		"total_price": TotalPrice,
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

// =========================================
type OrderRequest struct {
	TableNumber string `json:"table_number"`
	TotalPrice  int    `json:"total_price"`
}

func SubmitOrder(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "格式錯誤"})
		return
	}

	// 1. 建立訂單主表
	result, err := DB.Exec(`
		INSERT INTO orders (table_number, total_price,created_at)
		VALUES (?, ?,NOW())`, req.TableNumber, req.TotalPrice)
	if err != nil {
		c.JSON(500, gin.H{"error": "無法建立訂單"})
		return
	}

	orderID, _ := result.LastInsertId()

	// 2. 查詢該桌購物車商品
	rows, err := DB.Query(`
		SELECT product_id, product_name, quantity, total_price 
		FROM cart WHERE table_number = ?`, req.TableNumber)
	if err != nil {
		c.JSON(500, gin.H{"error": "讀取購物車失敗"})
		return
	}
	defer rows.Close()

	// 3. 寫入訂單細項
	for rows.Next() {
		var pid, qty, itemTotal int
		var name string
		rows.Scan(&pid, &name, &qty, &itemTotal)

		_, err = DB.Exec(`
			INSERT INTO order_items (order_id, product_id, product_name, quantity, total_price)
			VALUES (?, ?, ?, ?, ?)`, orderID, pid, name, qty, itemTotal)
		if err != nil {
			log.Println("寫入訂單細項失敗：", err)
			continue
		}
	}

	// 4. 清空該桌購物車
	_, err = DB.Exec(`DELETE FROM cart WHERE table_number = ?`, req.TableNumber)
	if err != nil {
		log.Println("清空購物車失敗：", err)
		c.JSON(500, gin.H{"error": "清空購物車失敗"})
		return
	}

	c.JSON(200, gin.H{"message": "訂單已送出", "order_id": orderID})
}

// ====================後端 Golang 實作 /add-batch-to-cart=================================
type CartItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type BatchCartRequest struct {
	TableNumber string     `json:"table_number"`
	Items       []CartItem `json:"items"`
}

func AddBatchToCart(c *gin.Context) {
	var req BatchCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "格式錯誤"})
		return
	}

	for _, item := range req.Items {
		_, err := DB.Exec(`
			INSERT INTO cart (product_id, product_name, quantity, total_price, table_number)
			SELECT ?, product_name, ?, price * ?, ?
			FROM products
			WHERE product_id = ?
			ON DUPLICATE KEY UPDATE 
				quantity = quantity + VALUES(quantity),
				total_price = total_price + VALUES(total_price)`,
			item.ProductID, item.Quantity, item.Quantity, req.TableNumber, item.ProductID)

		if err != nil {
			log.Println("批次加入購物車錯誤:", err)
			c.JSON(500, gin.H{"error": "資料庫寫入失敗"})
			return
		}
	}

	c.JSON(200, gin.H{"message": "商品已加入購物車"})
}
