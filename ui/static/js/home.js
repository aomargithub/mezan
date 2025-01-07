
 const copyLinks = document.querySelectorAll(`[id*="copy-a-"]`);
 copyLinks.forEach((link) => {
     link.addEventListener("click", handleClick);
 });
function handleClick(e) {
    e.stopPropagation()
    e.preventDefault()
    let shareId = e.target.getAttribute("data-share-id");
    navigator.clipboard.writeText("https://localhost:4000/mezanis/shareId/" + shareId);
}