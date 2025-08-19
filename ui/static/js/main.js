var navLinks = document.querySelectorAll("nav a");
for (var i = 0; i < navLinks.length; i++) {
	var link = navLinks[i]
	if (link.getAttribute('href') == window.location.pathname) {
		link.classList.add("live");
		break;
	}
}

// Delete confirmation using event listeners
function setupDeleteConfirmations() {
    var deleteButtons = document.querySelectorAll('button[title*="Delete"], button.btn-delete');
    
    for (var i = 0; i < deleteButtons.length; i++) {
        var button = deleteButtons[i];
        
        button.addEventListener('click', function(e) {
            var form = this.closest('form');
            if (form) {
                var action = form.getAttribute('action');
                
                var itemType = 'item';
                if (action.indexOf('/client/delete/') !== -1) {
                    itemType = 'client';
                } else if (action.indexOf('/project/delete/') !== -1) {
                    itemType = 'project';
                } else if (action.indexOf('/timesheet/delete/') !== -1) {
                    itemType = 'timesheet';
                } else if (action.indexOf('/invoice/delete/') !== -1) {
                    itemType = 'invoice';
                }
                
                var message = 'Are you sure you want to delete this ' + itemType + '?';
                
                if (!confirm(message)) {
                    e.preventDefault();
                    e.stopPropagation();
                    return false;
                }
            }
        });
    }
}

// Set up confirmations when page loads
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', setupDeleteConfirmations);
} else {
    setupDeleteConfirmations();
}