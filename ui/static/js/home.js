const copy = document.getElementById("copy-a");
copy.addEventListener("click", handleClick);
function handleClick(e) {
    let shareId = e.target.getAttribute("data-share-id");
    navigator.clipboard.writeText("https://localhost:4000/mezanis/shareId/" + shareId);
}