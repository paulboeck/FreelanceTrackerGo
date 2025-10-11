<?php
/**
 * Contact Form Submission Proxy for Freelance Tracker
 *
 * This script acts as a secure proxy between the contact form and Web3Forms API.
 * It keeps your Web3Forms access key hidden from public view.
 *
 * Setup:
 * 1. Get your access key from https://web3forms.com
 * 2. Replace 'YOUR_WEB3FORMS_ACCESS_KEY_HERE' below with your actual key
 * 3. Upload this file to your web server alongside index.html
 */

// Prevent direct access
if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    http_response_code(405);
    header('Content-Type: application/json');
    echo json_encode(['success' => false, 'message' => 'Method not allowed']);
    exit;
}

// CORS headers - restrict to your domain in production
// For testing, allow all origins. In production, change * to your actual domain
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: POST');
header('Access-Control-Allow-Headers: Content-Type');
header('Content-Type: application/json');

// Handle preflight OPTIONS request
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(200);
    exit;
}

// ============================================================================
// CONFIGURATION - CHANGE THIS!
// ============================================================================
define('WEB3FORMS_ACCESS_KEY', 'YOUR_WEB3FORMS_ACCESS_KEY_HERE');

// Optional: Restrict to specific domains (recommended for production)
// Uncomment and set your domain to prevent unauthorized use
// define('ALLOWED_ORIGIN', 'https://yourdomain.com');
// if (isset($_SERVER['HTTP_ORIGIN']) && $_SERVER['HTTP_ORIGIN'] !== ALLOWED_ORIGIN) {
//     http_response_code(403);
//     echo json_encode(['success' => false, 'message' => 'Unauthorized origin']);
//     exit;
// }

// ============================================================================
// VALIDATE ACCESS KEY IS SET
// ============================================================================
if (WEB3FORMS_ACCESS_KEY === 'YOUR_WEB3FORMS_ACCESS_KEY_HERE') {
    http_response_code(500);
    echo json_encode([
        'success' => false,
        'message' => 'Server configuration error: Web3Forms access key not set. Please update contact-submit.php'
    ]);
    exit;
}

// ============================================================================
// GET AND VALIDATE FORM DATA
// ============================================================================
$input = file_get_contents('php://input');
$data = json_decode($input, true);

if (!$data) {
    http_response_code(400);
    echo json_encode(['success' => false, 'message' => 'Invalid JSON data']);
    exit;
}

// Validate required fields
$required_fields = ['name', 'email', 'message'];
foreach ($required_fields as $field) {
    if (empty($data[$field])) {
        http_response_code(400);
        echo json_encode([
            'success' => false,
            'message' => "Missing required field: $field"
        ]);
        exit;
    }
}

// Basic email validation
if (!filter_var($data['email'], FILTER_VALIDATE_EMAIL)) {
    http_response_code(400);
    echo json_encode(['success' => false, 'message' => 'Invalid email address']);
    exit;
}

// Sanitize input data
$sanitized_data = [
    'name' => htmlspecialchars(strip_tags(trim($data['name'])), ENT_QUOTES, 'UTF-8'),
    'email' => filter_var(trim($data['email']), FILTER_SANITIZE_EMAIL),
    'message' => htmlspecialchars(strip_tags(trim($data['message'])), ENT_QUOTES, 'UTF-8')
];

// ============================================================================
// BASIC SPAM PROTECTION
// ============================================================================

// Check for honeypot field (if present in form)
if (isset($data['honeypot']) && !empty($data['honeypot'])) {
    // Silently reject spam
    http_response_code(200);
    echo json_encode(['success' => true, 'message' => 'Message sent successfully']);
    exit;
}

// Basic rate limiting (per IP)
session_start();
$ip = $_SERVER['REMOTE_ADDR'];
$last_submission_key = 'last_submission_' . md5($ip);

if (isset($_SESSION[$last_submission_key])) {
    $time_since_last = time() - $_SESSION[$last_submission_key];
    if ($time_since_last < 60) { // 60 seconds between submissions
        http_response_code(429);
        echo json_encode([
            'success' => false,
            'message' => 'Please wait a moment before submitting again'
        ]);
        exit;
    }
}

// ============================================================================
// FORWARD TO WEB3FORMS API
// ============================================================================

// Add the secret access key
$payload = [
    'access_key' => WEB3FORMS_ACCESS_KEY,
    'name' => $sanitized_data['name'],
    'email' => $sanitized_data['email'],
    'message' => $sanitized_data['message'],
    'from_name' => 'Freelance Tracker Contact Form',
    'subject' => 'New Contact Form Submission - Freelance Tracker'
];

// Initialize cURL
$ch = curl_init('https://api.web3forms.com/submit');
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/json',
    'Accept: application/json'
]);
curl_setopt($ch, CURLOPT_TIMEOUT, 10);

// Execute request
$response = curl_exec($ch);
$http_code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
$curl_error = curl_error($ch);
curl_close($ch);

// Handle cURL errors
if ($curl_error) {
    error_log("Web3Forms API Error: " . $curl_error);
    http_response_code(500);
    echo json_encode([
        'success' => false,
        'message' => 'Unable to send message. Please try again or email us directly.'
    ]);
    exit;
}

// Record successful submission for rate limiting
if ($http_code >= 200 && $http_code < 300) {
    $_SESSION[$last_submission_key] = time();
}

// Forward the response from Web3Forms
http_response_code($http_code);
echo $response;
?>
