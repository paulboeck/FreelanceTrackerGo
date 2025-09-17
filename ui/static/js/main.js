var navLinks = document.querySelectorAll("nav a");
for (var i = 0; i < navLinks.length; i++) {
	var link = navLinks[i]
	if (link.getAttribute('href') == window.location.pathname) {
		link.classList.add("live");
		break;
	}
}

// Custom confirmation dialog
function createConfirmationDialog() {
    var overlay = document.createElement('div');
    overlay.className = 'confirmation-overlay';
    overlay.innerHTML = `
        <div class="confirmation-dialog">
            <div class="confirmation-icon">
                <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
                </svg>
            </div>
            <h3 class="confirmation-title">Confirm Deletion</h3>
            <p class="confirmation-message"></p>
            <div class="confirmation-buttons">
                <button class="confirmation-btn confirmation-btn-cancel">Cancel</button>
                <button class="confirmation-btn confirmation-btn-confirm">Delete</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(overlay);
    return overlay;
}

function showConfirmation(message, onConfirm) {
    var overlay = document.querySelector('.confirmation-overlay');
    if (!overlay) {
        overlay = createConfirmationDialog();
    }
    
    var messageElement = overlay.querySelector('.confirmation-message');
    var cancelBtn = overlay.querySelector('.confirmation-btn-cancel');
    var confirmBtn = overlay.querySelector('.confirmation-btn-confirm');
    
    messageElement.textContent = message;
    
    // Show the dialog
    overlay.classList.add('show');
    
    // Focus the cancel button for accessibility
    setTimeout(function() {
        cancelBtn.focus();
    }, 100);
    
    // Handle cancel
    function handleCancel() {
        overlay.classList.remove('show');
        cleanup();
    }
    
    // Handle confirm
    function handleConfirm() {
        overlay.classList.remove('show');
        setTimeout(onConfirm, 300); // Wait for animation
        cleanup();
    }
    
    // Handle escape key
    function handleKeyDown(e) {
        if (e.key === 'Escape') {
            handleCancel();
        }
    }
    
    // Cleanup function
    function cleanup() {
        cancelBtn.removeEventListener('click', handleCancel);
        confirmBtn.removeEventListener('click', handleConfirm);
        overlay.removeEventListener('click', handleOverlayClick);
        document.removeEventListener('keydown', handleKeyDown);
    }
    
    // Handle clicking outside the dialog
    function handleOverlayClick(e) {
        if (e.target === overlay) {
            handleCancel();
        }
    }
    
    // Add event listeners
    cancelBtn.addEventListener('click', handleCancel);
    confirmBtn.addEventListener('click', handleConfirm);
    overlay.addEventListener('click', handleOverlayClick);
    document.addEventListener('keydown', handleKeyDown);
}

// Delete confirmation using event listeners
function setupDeleteConfirmations() {
    var deleteButtons = document.querySelectorAll('button[title*="Delete"], button.btn-delete');
    
    for (var i = 0; i < deleteButtons.length; i++) {
        var button = deleteButtons[i];
        
        button.addEventListener('click', function(e) {
            e.preventDefault();
            e.stopPropagation();
            
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
                
                var message = 'Are you sure you want to delete this ' + itemType + '? This action cannot be undone.';
                
                showConfirmation(message, function() {
                    form.submit();
                });
            }
            
            return false;
        });
    }
}

// Toggle client details functionality
function toggleClientDetails() {
    var details = document.getElementById('client-details');
    var icon = document.getElementById('toggle-icon');
    
    if (!details || !icon) return;
    
    if (details.classList.contains('hidden')) {
        details.classList.remove('hidden');
        icon.textContent = '▼';
    } else {
        details.classList.add('hidden');
        icon.textContent = '▶';
    }
}

// Set up toggle button when page loads
function setupClientDetailsToggle() {
    var toggleButton = document.getElementById('toggle-details');
    if (toggleButton) {
        toggleButton.addEventListener('click', toggleClientDetails);
    }
}

// Email progress functionality
function showEmailProgress() {
    var modal = document.getElementById('emailProgressModal');
    if (!modal) return;
    
    // Reset all steps
    var steps = modal.querySelectorAll('.progress-step');
    var progressFill = modal.querySelector('#progressFill');
    
    steps.forEach(function(step) {
        step.classList.remove('active', 'complete');
    });
    progressFill.style.width = '0%';
    
    // Show modal
    modal.classList.add('show');
    
    // Start step 1: PDF generation
    setTimeout(function() {
        document.getElementById('step-pdf').classList.add('active');
        progressFill.style.width = '33%';
    }, 100);
    
    // Start step 2: Email sending (after 1.5 seconds)
    setTimeout(function() {
        document.getElementById('step-pdf').classList.remove('active');
        document.getElementById('step-pdf').classList.add('complete');
        document.getElementById('step-email').classList.add('active');
        progressFill.style.width = '66%';
    }, 1500);
    
    // Complete (after another 1.5 seconds)
    setTimeout(function() {
        document.getElementById('step-email').classList.remove('active');
        document.getElementById('step-email').classList.add('complete');
        document.getElementById('step-complete').classList.add('active');
        document.getElementById('step-complete').classList.add('complete');
        progressFill.style.width = '100%';
        
        // Hide modal after showing complete state for 1 second
        setTimeout(function() {
            modal.classList.remove('show');
        }, 1000);
    }, 3000);
}

function setupEmailProgressButtons() {
    var emailForms = document.querySelectorAll('form[action*="/invoice/email/"]');
    
    emailForms.forEach(function(form) {
        form.addEventListener('submit', function(e) {
            var button = form.querySelector('.btn-email');
            if (button) {
                button.classList.add('loading');
                showEmailProgress();
            }
        });
    });
}

// Set up all functionality when page loads
function setupPageFunctions() {
    setupDeleteConfirmations();
    setupClientDetailsToggle();
    setupEmailProgressButtons();
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', setupPageFunctions);
} else {
    setupPageFunctions();
}