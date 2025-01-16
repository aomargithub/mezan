const shareType = document.getElementById("shareType");
const share = document.getElementById("share");
const amount = document.getElementById("amount");
const amountDisplay = document.getElementById("amountDisplay");
const expenseItemTotalAmount = document.getElementById("expenseItemTotalAmount");

shareType.addEventListener("change", handleShareTypeChange);
share.addEventListener("input", handleShareChange);

let shareTypeValue = shareType.value, shareValue = share.value, expenseItemTotalAmountValue = expenseItemTotalAmount.value;
function handleShareTypeChange(e) {
    shareTypeValue = e.target.value;
    calculate();
}

function handleShareChange(e) {
    shareValue = e.target.value;
    calculate();
}

function calculate() {
    if (shareTypeValue === 'PERCENTAGE') {
        amount.value = (expenseItemTotalAmountValue * shareValue) / 100;
        amountDisplay.value = (expenseItemTotalAmountValue * shareValue) / 100;
    } else {
        amount.value =  shareValue;
        amountDisplay.value =  shareValue;
    }
}