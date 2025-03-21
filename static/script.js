async function addToCart(productId) {
    const tableNumber = document.getElementById('tableSelect').value;

    const res = await fetch('http://localhost:8080/add-to-cart', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ product_id: productId, table_number: tableNumber })
    });

    const data = await res.json();
    alert(data.message || data.error);
  }

  async function loadCart() {
    const tableNumber = document.getElementById('tableSelect').value;
    const res = await fetch(`http://localhost:8080/get-cart?table_number=${tableNumber}`);
    const data = await res.json();

    const cartDiv = document.getElementById('cart');
    const totalDiv = document.getElementById('total');

    cartDiv.innerHTML = '';
    totalDiv.innerHTML = '';

    if (data.error) {
      cartDiv.textContent = '錯誤：' + data.error;
      return;
    }

    data.cart.forEach(item => {
      const p = document.createElement('p');
      p.textContent = item;
      cartDiv.appendChild(p);
    });

    totalDiv.textContent = '總價為 ' + data.total_price + ' 元';
  }

  async function clearCart() {
    const tableNumber = document.getElementById('tableSelect').value;

    const res = await fetch(`http://localhost:8080/clear-cart?table_number=${tableNumber}`, {
      method: 'DELETE'
    });

    const data = await res.json();
    alert(data.message || data.error);

    loadCart(); // 重新載入購物車畫面
  }

  //這邊功能是漢堡標籤、蛋餅、飲料，一開始先隱藏
  function showCategory(category) {
    // 隱藏所有類別
    document.querySelectorAll('.category').forEach(div => {
      div.style.display = 'none';
    });

    // 顯示被點選的類別
    document.getElementById(category).style.display = 'block';
  }

  let cart = {}; // 用來存每個商品的數量，例如:cart={1: 2, 2: 3}商品ID 1豬排蛋漢堡有兩個 商品ID 2 雞排蛋漢堡有3個

  function updateQuantity(productId, change) {
    const tableNumber = document.getElementById('tableSelect').value;

    // 確保商品有初始數量
    if (!cart[productId]) {
      cart[productId] = 0;
    }

    // 計算新的數量（避免變成負數）
    let newQuantity = cart[productId] + change;
    if (newQuantity < 0) newQuantity = 0;//如果數值為負數，把newQuantity設為0

    cart[productId] = newQuantity;

    // 更新畫面顯示數量
    document.getElementById(`qty-${productId}`).textContent = newQuantity;

    // 當 + 按鈕被按下時，新增商品
    if (change > 0) {
      fetch("http://localhost:8080/add-to-cart", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ product_id: productId, table_number: tableNumber })
      }).then(res => res.json()).then(data => console.log(data));
    }else if (change < 0 && newQuantity > 0) {//當 - 按鈕被按下時，數量還大於0，更新商品數量
      fetch("http://localhost:8080/update-cart", {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ product_id: productId, table_number: tableNumber, quantity: newQuantity })
      }).then(res => res.json()).then(data => console.log("PUT /update-cart 回應：", data));
    }else if (change < 0 && newQuantity === 0) {//如果數量變成0，則刪除該商品
      fetch(`http://localhost:8080/remove-from-cart?product_id=${productId}&table_number=${tableNumber}`, {
        method: "DELETE"
      }).then(res => res.json()).then(data => console.log(data));
    }
  }