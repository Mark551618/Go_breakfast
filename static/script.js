document.addEventListener("DOMContentLoaded", function () {
    const API_URL = "http://localhost:8080"; // 你的 API 伺服器網址
    let tableNumber = "A1"; // 預設桌號，可由用戶選擇

    document.addEventListener("DOMContentLoaded", function() {
        fetchProducts(); // 頁面加載後自動執行
    });
    
    function fetchProducts() {
        console.log("🔍 正在載入商品...");
        
        fetch("http://localhost:8080/products")
            .then(response => {
                if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
                return response.json();
            })
            .then(data => {
                console.log("✅ 商品列表:", data);
                const productList = document.getElementById("product-list");
                productList.innerHTML = ""; // 清空舊內容
                data.forEach(product => {
                    const productItem = document.createElement("div");
                    productItem.classList.add("product-item");
                    productItem.innerHTML = `
                        <h3>${product.product_name}</h3>
                        <p>價格: $${product.price}</p>
                        <button onclick="addToCart(${product.product_id})">加入購物車</button>
                    `;
                    productList.appendChild(productItem);
                });
            })
            .catch(error => console.error("❌ 獲取商品失敗:", error));
    }
    
    

    // **新增商品到購物車**
    function addToCart(productID) {
        fetch(`${API_URL}/cart/add`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ product_id: productID, table_number: tableNumber })
        })
            .then(response => response.json())
            .then(data => {
                alert(data.message);
                fetchCart(); // 更新購物車顯示
            })
            .catch(error => console.error("加入購物車失敗:", error));
    }

    // **獲取購物車內容**
    function fetchCart() {
        fetch(`${API_URL}/cart?table_number=${tableNumber}`)
            .then(response => response.json())
            .then(data => {
                const cartList = document.getElementById("cart-list");
                const totalPrice = document.getElementById("total-price");
                cartList.innerHTML = ""; // 清空購物車列表
                let total = 0;
                data.cart.forEach(item => {
                    total += item.price * item.quantity;
                    const cartItem = document.createElement("div");
                    cartItem.classList.add("cart-item");
                    cartItem.innerHTML = `
                        <h4>${item.product_name}</h4>
                        <p>價格: $${item.price} x ${item.quantity}</p>
                        <button onclick="updateCart(${item.product_id}, ${item.quantity + 1})">+</button>
                        <button onclick="updateCart(${item.product_id}, ${item.quantity - 1})">-</button>
                        <button onclick="removeFromCart(${item.product_id})">刪除</button>
                    `;
                    cartList.appendChild(cartItem);
                });
                totalPrice.textContent = `總計: $${total}`;
            })
            .catch(error => console.error("獲取購物車失敗:", error));
    }

    // **刪除購物車內商品**
    function removeFromCart(productID) {
        fetch(`${API_URL}/cart/remove?table_number=${tableNumber}&product_id=${productID}`, {
            method: "POST"
        })
            .then(response => response.json())
            .then(data => {
                alert(data.message);
                fetchCart(); // 更新購物車顯示
            })
            .catch(error => console.error("刪除商品失敗:", error));
    }

    // **清空購物車**
    function clearCart() {
        fetch(`${API_URL}/cart/clear?table_number=${tableNumber}`, {
            method: "POST"
        })
            .then(response => response.json())
            .then(data => {
                alert(data.message);
                fetchCart(); // 更新購物車顯示
            })
            .catch(error => console.error("清空購物車失敗:", error));
    }

    // **更新購物車數量**
    function updateCart(productID, quantity) {
        if (quantity <= 0) {
            removeFromCart(productID);
            return;
        }
        fetch(`${API_URL}/cart/update?table_number=${tableNumber}&product_id=${productID}`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ quantity: quantity })
        })
            .then(response => response.json())
            .then(data => {
                alert(data.message);
                fetchCart(); // 更新購物車顯示
            })
            .catch(error => console.error("更新購物車失敗:", error));
    }

    // **結帳**
    function checkout() {
        fetch(`${API_URL}/checkout?table_number=${tableNumber}`, {
            method: "POST"
        })
            .then(response => response.json())
            .then(data => {
                alert(data.message);
                fetchCart(); // 清空購物車顯示
            })
            .catch(error => console.error("結帳失敗:", error));
    }

    // **初始化**
    fetchProducts();
    fetchCart();

    // **綁定按鈕**
    document.getElementById("clear-cart").addEventListener("click", clearCart);
    document.getElementById("checkout").addEventListener("click", checkout);
});
