function formatLocal(date) {
    const pad = n => String(n).padStart(2, '0');
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ` +
        `${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

export function localizeTimes() {
    document.querySelectorAll('.local-time-popover time[datetime]').forEach(el => {
        const d = new Date(el.getAttribute('datetime'));
        el.closest('.local-time-popover').title = formatLocal(d);
    });
}
