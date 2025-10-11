/**
 * Contact Form Handler for Freelance Tracker
 * Handles form validation and submission
 */

document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('contactForm');
    const formStatus = document.getElementById('formStatus');

    if (!form) return; // Exit if form doesn't exist on this page

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        // Get form data
        const formData = {
            name: document.getElementById('name').value.trim(),
            email: document.getElementById('email').value.trim(),
            message: document.getElementById('message').value.trim()
        };

        // Basic validation
        if (!formData.name || !formData.email || !formData.message) {
            showStatus('Please fill in all fields.', 'error');
            return;
        }

        if (!isValidEmail(formData.email)) {
            showStatus('Please enter a valid email address.', 'error');
            return;
        }

        // Disable submit button and show loading state
        const submitButton = form.querySelector('button[type="submit"]');
        const originalButtonText = submitButton.textContent;
        submitButton.disabled = true;
        submitButton.textContent = 'Sending...';

        try {
            // Submit to PHP proxy which securely forwards to Web3Forms
            const response = await fetch('contact-submit.php', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(formData)
            });

            const result = await response.json();

            if (response.ok && result.success) {
                // Success!
                showStatus('Message sent successfully! I\'ll get back to you soon.', 'success');

                // Reset form
                form.reset();

                // Clear status after 5 seconds
                setTimeout(() => {
                    formStatus.textContent = '';
                }, 5000);
            } else {
                // Server returned an error
                const errorMessage = result.message || 'Unable to send message. Please try again.';
                showStatus(errorMessage, 'error');
            }

        } catch (error) {
            console.error('Form submission error:', error);

            // Network error or server unreachable
            showStatus(
                'Unable to send message. Please email us directly at info@small-biz-software.com',
                'error'
            );
        } finally {
            // Re-enable submit button
            submitButton.disabled = false;
            submitButton.textContent = originalButtonText;
        }
    });

    function showStatus(message, type) {
        formStatus.textContent = message;
        formStatus.className = 'form-status ' + type;
    }

    function isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }
});
