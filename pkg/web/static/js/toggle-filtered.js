export function toggleFilteredStatus() {
    const toggleButton = document.getElementById("toggle-filtered-alerts");
    const filteredTable = document.getElementById("filtered-table");

    toggleButton.addEventListener("click", () => {
        filteredTable.classList.toggle("hidden");
        const filteredAreShown = !filteredTable.classList.contains("hidden");
        toggleButton.innerText = filteredAreShown ? "Hide" : "Show";
        localStorage.setItem("filteredAreShown", filteredAreShown.toString());
    });

    if (localStorage.getItem("filteredAreShown") === "true") {
        toggleButton.click();
    }
}
