const shareType = document.getElementById("shareType");
const share = document.getElementById("share");
const amount = document.getElementById("amount");
const amountDisplay = document.getElementById("amountDisplay");
const expenseTotalAmount = document.getElementById("expenseTotalAmount");

shareType.addEventListener("change", handleShareTypeChange);
share.addEventListener("input", handleShareChange);

let shareTypeValue = shareType.value, shareValue = share.value, expenseTotalAmountValue = expenseTotalAmount.value;
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
        amount.value = (expenseTotalAmountValue * shareValue) / 100;
        amountDisplay.value = (expenseTotalAmountValue * shareValue) / 100;
    } else {
        amount.value =  shareValue;
        amountDisplay.value =  shareValue;
    }
}