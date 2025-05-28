// Main JavaScript file for Frigate Alerter Web UI

document.addEventListener('DOMContentLoaded', function() {
    // Enable tooltips everywhere
    const tooltips = document.querySelectorAll('[data-bs-toggle="tooltip"]');
    tooltips.forEach(tooltip => {
        new bootstrap.Tooltip(tooltip);
    });

    // Format all dates to local timezone
    document.querySelectorAll('time.format-date').forEach(timeElement => {
        const timestamp = timeElement.getAttribute('datetime');
        if (timestamp) {
            const date = new Date(timestamp);
            timeElement.textContent = date.toLocaleString();
        }
    });

    // Active link highlighting in navbar
    const currentPath = window.location.pathname;
    document.querySelectorAll('.navbar-nav .nav-link').forEach(link => {
        if (link.getAttribute('href') === currentPath) {
            link.classList.add('active');
        }
    });

    // Handle notification fadeout
    function setupNotificationFadeout(notificationElement) {
        if (notificationElement && !notificationElement.classList.contains('d-none')) {
            setTimeout(() => {
                notificationElement.classList.add('fade-out');
                setTimeout(() => {
                    notificationElement.classList.add('d-none');
                    notificationElement.classList.remove('fade-out');
                }, 500);
            }, 5000);
        }
    }

    // Setup for any notification on the page
    const notifications = document.querySelectorAll('.alert-notification');
    notifications.forEach(setupNotificationFadeout);
});
