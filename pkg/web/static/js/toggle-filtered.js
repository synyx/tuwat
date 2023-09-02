export function toggleFilteredStatus() {
    const toggleButton = document.getElementById("toggle-filtered-alerts");
    const filteredTable = document.getElementById("filtered-table");

    toggleButton.addEventListener("click", () => {
        filteredTable.classList.toggle("hidden");
        const filteredAreShown = !filteredTable.classList.contains("hidden");
        toggleButton.innerText = filteredAreShown ? "Hide" : "Show";
        localStorage.setItem("filteredAreShown", filteredAreShown.toString());
    });

    function reRegisterToggleFilteredStatus(event) {
        const fallbackToDefaultActions = event.detail.render

        event.detail.render = function (streamElement) {
            fallbackToDefaultActions(streamElement)
            toggleFilteredStatus();
        }
    }
    document.addEventListener("turbo:before-stream-render", reRegisterToggleFilteredStatus);

    if (localStorage.getItem("filteredAreShown") === "true") {
        toggleButton.click();
    }
}
