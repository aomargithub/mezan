const quantity = document.getElementById("quantity");
const amount = document.getElementById("amount");
const totalAmount = document.getElementById("totalAmount");
quantity.addEventListener("input", handleQuantityChange);
amount.addEventListener("input", handleAmountChange);
function handleQuantityChange(e) {
    let quantity = e.target.value;
    totalAmount.value = quantity * amount.value;
}

function handleAmountChange(e) {
    let amount = e.target.value;
    totalAmount.value = amount * quantity.value;
}