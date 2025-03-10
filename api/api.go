package api

import (
	"breakfast-shop/mysql"
	//"database/sql"
	//"fmt"
	//"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	//"github.com/go-sql-driver/mysql" //  這行是關鍵，讓 Go 能辨識 `mysql.MySQLError`
)

// 獲取所有商品
func GetProducts(c *gin.Context) {
	rows, err := mysql.DB.Query("SELECT product_id, product_name, price FROM products")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var product_id, price int
		var product_name string
		rows.Scan(&product_id, &product_name, &price)
		products = append(products, gin.H{"product_id": product_id, "product_name": product_name, "price": price})
	}
	c.JSON(200, products)
}

// 新增商品到購物車內
func AddToCart(c *gin.Context) {
	var req struct {
		ProductID   int    `json:"product_id"`
		TableNumber string `json:"table_number"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// 檢查 `products` 是否存在此商品
	var exists int
	err := mysql.DB.QueryRow("SELECT COUNT(*) FROM products WHERE product_id = ?", req.ProductID).Scan(&exists)
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}
	if exists == 0 {
		c.JSON(404, gin.H{"error": "Product not found"})
		return
	}

	// 插入或更新購物車
	_, err = mysql.DB.Exec(`
		INSERT INTO cart (product_id, product_name, quantity, price, table_number)
		SELECT ?, product_name, 1, price, ?
		FROM products
		WHERE product_id = ?
		ON DUPLICATE KEY UPDATE 
			quantity = quantity + 1,
			product_name = VALUES(product_name),
			price = VALUES(price)`, req.ProductID, req.TableNumber, req.ProductID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Product added to cart"})
}

// 取得購物車內的商品資訊
func GetCart(c *gin.Context) {
	// 🔹 從 `table_number` 參數取得桌號
	tableNumber := c.Query("table_number")
	if tableNumber == "" {
		c.JSON(400, gin.H{"error": "Missing table_number"})
		return
	}

	// 🔹 查詢購物車內容
	rows, err := mysql.DB.Query(`
		SELECT product_id, product_name, quantity, price 
		FROM cart 
		WHERE table_number = ?`, tableNumber)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var cart []map[string]interface{}
	var totalPrice int

	// 🔹 迭代每一行的查詢結果
	for rows.Next() {
		var productID, price, quantity int
		var productName string
		err := rows.Scan(&productID, &productName, &quantity, &price)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error reading cart data"})
			return
		}

		// 計算總價
		totalPrice += price * quantity

		// 加入購物車結果陣列
		cart = append(cart, gin.H{
			"product_id":   productID,
			"product_name": productName,
			"quantity":     quantity,
			"price":        price,
			"total":        price * quantity, // 單項商品的總價格
		})
	}

	// 🔹 返回 JSON 給前端
	c.JSON(200, gin.H{
		"cart":        cart,
		"total_price": totalPrice,
	})
}

// 刪除購物車內容
func RemoveFromCart(c *gin.Context) {
	tableNumber := c.Query("table_number")
	productID := c.Query("product_id")

	if tableNumber == "" || productID == "" {
		c.JSON(400, gin.H{"error": "Missing table_number or product_id"})
		return
	}

	productIDInt, err := strconv.Atoi(productID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid product_id"})
		return
	}

	result, err := mysql.DB.Exec(`DELETE FROM cart WHERE table_number = ? AND product_id = ?`, tableNumber, productIDInt)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Product not found in cart"})
		return
	}

	c.JSON(200, gin.H{"message": "Product removed from cart"})
}

// 清除整個購物車
func ClearCart(c *gin.Context) {
	// 🔹 從 `table_number` 參數取得桌號
	tableNumber := c.Query("table_number")
	if tableNumber == "" {
		c.JSON(400, gin.H{"error": "Missing table_number"})
		return
	}

	// 🔹 刪除該桌的所有購物車內容
	_, err := mysql.DB.Exec(`DELETE FROM cart WHERE table_number = ?`, tableNumber)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 🔹 返回成功訊息
	c.JSON(200, gin.H{
		"message":      "Cart cleared successfully",
		"table_number": tableNumber,
	})
}

// 確認訂單
func UpdateCart(c *gin.Context) {
	// 🔹 取得 `table_number` 和 `product_id`
	tableNumber := c.Query("table_number")
	if tableNumber == "" {
		c.JSON(400, gin.H{"error": "Missing table_number"})
		return
	}

	productID := c.Query("product_id")
	if productID == "" {
		c.JSON(400, gin.H{"error": "Missing product_id"})
		return
	}

	// 🔹 解析 JSON body，取得新的 `quantity`
	var req struct {
		Quantity int `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON request"})
		return
	}

	// 🔹 如果 `quantity == 0`，則刪除該商品
	if req.Quantity == 0 {
		_, err := mysql.DB.Exec(`DELETE FROM cart WHERE table_number = ? AND product_id = ?`, tableNumber, productID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"message":      "Product removed from cart",
			"product_id":   productID,
			"table_number": tableNumber,
		})
		return
	}

	// 🔹 更新購物車數量
	result, err := mysql.DB.Exec(`
		UPDATE cart 
		SET quantity = ? 
		WHERE table_number = ? AND product_id = ?`, req.Quantity, tableNumber, productID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 🔹 檢查是否有更新到資料
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(500, gin.H{"error": "Error retrieving update result"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Product not found in cart"})
		return
	}

	// 🔹 回傳成功訊息
	c.JSON(200, gin.H{
		"message":      "Cart updated successfully",
		"product_id":   productID,
		"quantity":     req.Quantity,
		"table_number": tableNumber,
	})
}

// 結帳
func Checkout(c *gin.Context) {
	tableNumber := c.Query("table_number")
	if tableNumber == "" {
		c.JSON(400, gin.H{"error": "Missing table_number"})
		return
	}

	rows, err := mysql.DB.Query(`SELECT product_id, product_name, quantity, price FROM cart WHERE table_number = ?`, tableNumber)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var totalPrice int
	var cartItems []struct {
		ProductID   int
		ProductName string
		Quantity    int
		Price       int
	}

	for rows.Next() {
		var item struct {
			ProductID   int
			ProductName string
			Quantity    int
			Price       int
		}
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.Quantity, &item.Price); err != nil {
			c.JSON(500, gin.H{"error": "Error reading cart data"})
			return
		}
		cartItems = append(cartItems, item)
		totalPrice += item.Price * item.Quantity
	}

	if len(cartItems) == 0 {
		c.JSON(400, gin.H{"error": "Cart is empty, cannot checkout"})
		return
	}

	result, _ := mysql.DB.Exec(`INSERT INTO orders (table_number, total_price) VALUES (?, ?)`, tableNumber, totalPrice)

	orderID, _ := result.LastInsertId()

	for _, item := range cartItems {
		_, err := mysql.DB.Exec(`INSERT INTO order_items (order_id, product_id, product_name, quantity, price) 
			VALUES (?, ?, ?, ?, ?)`, orderID, item.ProductID, item.ProductName, item.Quantity, item.Price)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error inserting order items"})
			return
		}
	}

	_, _ = mysql.DB.Exec(`DELETE FROM cart WHERE table_number = ?`, tableNumber)

	c.JSON(200, gin.H{"message": "Order placed successfully", "order_id": orderID, "total_price": totalPrice})
}
