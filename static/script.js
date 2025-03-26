//======================================================================================================
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
//=======================================================================================================
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
//=========================================================================================================
async function clearCart() {
  const tableNumber = document.getElementById('tableSelect').value;

  const res = await fetch(`http://localhost:8080/clear-cart?table_number=${tableNumber}`, {
    method: 'DELETE'
  });

  const data = await res.json();
  alert(data.message || data.error);

  loadCart(); // 重新載入購物車畫面
}
//=============================================================================================================
//這邊功能是漢堡標籤、蛋餅、飲料，一開始先隱藏
function showCategory(category) {
  // 隱藏所有類別
  document.querySelectorAll('.category').forEach(div => {
    div.style.display = 'none';
  });

  // 顯示被點選的類別
  document.getElementById(category).style.display = 'block';
}
//================================updateQuantity() 只更新前端，不呼叫後端=========================================
let cart = {}; // 用來存每個商品的數量，例如:cart={1: 2, 2: 3}商品ID 1豬排蛋漢堡有兩個 商品ID 2 雞排蛋漢堡有3個

function updateQuantity(productId, change) {
  if (!cart[productId]) cart[productId] = 0;

  let newQty = cart[productId] + change;
  if (newQty < 0) newQty = 0;
  cart[productId] = newQty;

  document.getElementById(`qty-${productId}`).textContent = newQty;
}
//================"送出按鍵"加上 submitCart() 方法，把暫存的 cart 傳到後端===================================================
async function submitCart() {
  const tableNumber = document.getElementById('tableSelect').value;
  
  // 將 cart 轉為要送的格式
  const payload = {
    table_number: tableNumber,
    items: []
  };

  for (let productId in cart) {
    const quantity = cart[productId];
    if (quantity > 0) {
      payload.items.push({
        product_id: parseInt(productId),
        quantity: quantity
      });
    }
  }

  if (payload.items.length === 0) {
    alert("請先選擇商品再送出！");
    return;
  }

  const res = await fetch("http://localhost:8080/add-batch-to-cart", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  });

  const data = await res.json();
  alert(data.message || "送出成功");

  // 清除暫存 + 畫面
  cart = {};
  payload.items.forEach(item => {
    document.getElementById(`qty-${item.product_id}`).textContent = 0;
  });
}

//=================================訂單送出的功能=================================================================
async function submitOrder() {
  const tableNumber = document.getElementById('tableSelect').value;
  const res = await fetch(`http://localhost:8080/get-cart?table_number=${tableNumber}`);
  const data = await res.json();

  if (data.error) {
    alert('無法取得購物車資料');
    return;
  }

  const orderData = {
    table_number: tableNumber
  };

  const submitRes = await fetch("http://localhost:8080/submit-order", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(orderData)
  });

  const result = await submitRes.json();

  //  顯示訂單號碼與總金額
  alert(`訂單送出成功！\n訂單編號：${result.order_id}\n總金額：${result.amount_of_money} 元`);

  //  清空畫面
  document.getElementById('cart').innerHTML = '';
  document.getElementById('total').textContent = '';
  for (let i = 1; i <= 9; i++) {
    const qtySpan = document.getElementById(`qty-${i}`);
    if (qtySpan) qtySpan.textContent = 0;
  }

  //  清空 JS 內部暫存
  if (typeof cart === 'object') {
    cart = {};
  }
}


