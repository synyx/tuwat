export function toggleFilteredStatus(button, table) {
    const toggleButton = document.getElementById(button);
    const filteredTable = document.getElementById(table);

    toggleButton.addEventListener("click", () => {
        filteredTable.classList.toggle("hidden");
        const filteredAreShown = !filteredTable.classList.contains("hidden");
        toggleButton.innerText = filteredAreShown ? "Hide" : "Show";
        localStorage.setItem(button+"AreShown", filteredAreShown.toString());
        if (filteredAreShown) {
            filteredTable.scrollIntoView();
        }
    });

    function reRegisterToggleFilteredStatus(event) {
        const fallbackToDefaultActions = event.detail.render

        event.detail.render = function (streamElement) {
            fallbackToDefaultActions(streamElement)
            toggleFilteredStatus();
        }
    }
    document.addEventListener("turbo:before-stream-render", reRegisterToggleFilteredStatus);

    if (localStorage.getItem(button+"AreShown") === "true") {
        filteredTable.classList.remove("hidden");
        toggleButton.innerText = "Show";
    }
}
