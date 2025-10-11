# Freelance Tracker Landing Page

This folder contains the marketing landing page for Freelance Tracker, separate from the main application.

## Files

- **index.html** - Main landing page with hero, features, pricing, and contact sections
- **get-started.html** - Payment instructions and next steps for customers
- **styles.css** - Professional, trustworthy styling with blue/gray color palette
- **contact-form.js** - Contact form validation and submission handling
- **contact-submit.php** - Secure PHP proxy for Web3Forms API (keeps access key hidden)

## Deployment

These are static HTML/CSS/JS files that can be hosted on any web server, including Namecheap shared hosting.

### To Deploy to Namecheap:

1. Upload all files to your `public_html` directory (or subdirectory)
2. Access via your domain (e.g., https://yourdomain.com)

## Setup Required

### 1. Update Payment Information (REQUIRED)

In `get-started.html`, replace the placeholder payment handles:

```html
<!-- Line ~58 -->
<p class="payment-handle">@YourVenmoHandle</p>

<!-- Line ~70 -->
<p class="payment-handle">$YourCashAppTag</p>
```

### 2. Configure Contact Form (REQUIRED)

The contact form uses a secure PHP proxy to hide your Web3Forms API key. Setup is simple:

#### Step 1: Get Your Web3Forms Access Key

1. Sign up at https://web3forms.com (free tier: 250 submissions/month)
2. Verify your email address
3. Copy your Access Key from the dashboard

#### Step 2: Update the PHP Script

Open `contact-submit.php` and find line 33:

```php
define('WEB3FORMS_ACCESS_KEY', 'YOUR_WEB3FORMS_ACCESS_KEY_HERE');
```

Replace `YOUR_WEB3FORMS_ACCESS_KEY_HERE` with your actual access key:

```php
define('WEB3FORMS_ACCESS_KEY', 'abc123-your-actual-key-here');
```

#### Step 3: Configure Web3Forms Settings (Recommended)

In your Web3Forms dashboard:

1. **Set notification email** to `info@small-biz-software.com`
2. **Enable domain whitelisting** - Add your domain (e.g., `yourdomain.com`)
   - This prevents unauthorized use of your form
3. **Enable spam filtering** for extra protection

#### Step 4: Optional - Restrict to Your Domain

For production, uncomment lines 40-45 in `contact-submit.php`:

```php
define('ALLOWED_ORIGIN', 'https://yourdomain.com');
if (isset($_SERVER['HTTP_ORIGIN']) && $_SERVER['HTTP_ORIGIN'] !== ALLOWED_ORIGIN) {
    http_response_code(403);
    echo json_encode(['success' => false, 'message' => 'Unauthorized origin']);
    exit;
}
```

Replace `https://yourdomain.com` with your actual domain.

#### Security Features Built-In

The PHP proxy includes:
- ✅ **Hidden API key** - Never exposed to the browser
- ✅ **Input validation** - Validates required fields and email format
- ✅ **Input sanitization** - Strips harmful HTML/scripts
- ✅ **Rate limiting** - 60 seconds between submissions per IP
- ✅ **Honeypot support** - Catches bot submissions
- ✅ **CORS protection** - Can restrict to your domain only

#### How It Works

```
User fills form → JavaScript validates → POST to contact-submit.php
→ PHP validates & sanitizes → Forwards to Web3Forms API
→ Response back to user
```

Your API key stays safely on the server and is never visible in the browser.

## Automated Payment Solutions

### Current Setup: Manual Process

The current `get-started.html` page uses a manual process:
1. Customer sends payment via Venmo/Cash App
2. You receive payment notification
3. You manually email the download link

**Limitations:**
- Venmo and Cash App don't offer public APIs for automated payment verification
- You must manually process each order
- Customers wait for your response

### Recommended: Automated Solutions

For a fully automated experience (payment → instant download), consider these alternatives:

#### Option 1: Gumroad (Best for Digital Products)

**Pros:**
- Built for selling digital downloads
- Automatic delivery of files after payment
- Accepts credit cards, Apple Pay, Google Pay
- Handles EU VAT automatically
- Professional checkout experience
- Free to start (8.5% + 30¢ per transaction)

**Cons:**
- Takes a percentage of each sale
- Not Venmo/Cash App

**Setup:**
1. Sign up at https://gumroad.com
2. Create a product and upload your application
3. Replace the "Get Started" button with Gumroad's checkout link

**Implementation:**
```html
<!-- In index.html and get-started.html -->
<a href="https://yourname.gumroad.com/l/freelance-tracker" class="btn-primary">
    Get Started - $49
</a>
```

#### Option 2: Stripe Payment Links (Most Professional)

**Pros:**
- Professional payment processing
- Automatic email delivery with custom links
- Accept all major credit cards
- Lower fees than Gumroad (2.9% + 30¢)
- Trusted by customers
- Can integrate with webhooks for automation

**Cons:**
- Requires more technical setup
- Must handle download hosting/delivery
- Not Venmo/Cash App

**Setup:**
1. Sign up at https://stripe.com
2. Create a Payment Link product
3. Set up email delivery or use webhooks to trigger download emails
4. Update buttons to link to Stripe checkout

#### Option 3: PayPal Buy Now Buttons

**Pros:**
- Widely recognized and trusted
- Can set up automated download delivery
- Reasonable fees (2.9% + 30¢)
- IPN (Instant Payment Notification) for automation

**Cons:**
- Requires IPN setup for automation
- PayPal branding
- Not Venmo/Cash App

**Setup:**
1. Create a PayPal business account
2. Generate a Buy Now button
3. Set up IPN to trigger download emails
4. Embed button on your page

#### Option 4: Hybrid Approach

Keep Venmo/Cash App for customers who prefer it, but add a "Credit Card" option using Stripe or Gumroad for those who want instant access:

```html
<div class="payment-options">
    <div class="payment-method">
        <h3>Instant Access</h3>
        <p>Pay with credit card</p>
        <a href="https://yourname.gumroad.com/l/freelance-tracker" class="btn-primary">
            Buy Now - $49
        </a>
        <p class="payment-note">Instant download link</p>
    </div>
    <div class="payment-method">
        <h3>Venmo/Cash App</h3>
        <p>Manual processing</p>
        <a href="get-started.html" class="btn-secondary">
            Pay via Venmo/Cash App
        </a>
        <p class="payment-note">Link sent within 24hrs</p>
    </div>
</div>
```

### Download Delivery Options

Once you have automated payments, you need to deliver the download:

1. **Gumroad** - Hosts files for you (easiest)
2. **Cloud Storage Link** - Generate time-limited download links (DigitalOcean Spaces, AWS S3)
3. **Email with Webhook** - Use Stripe webhook → send email with download link
4. **Download Portal** - Create a simple web page with unique access codes

## Next Steps

1. ⏳ **Update Venmo/Cash App handles** in `get-started.html` (lines 58 & 70)
2. ⏳ **Configure contact form** - Add Web3Forms access key to `contact-submit.php` (line 33)
3. ⏳ **Upload to Namecheap** - Upload ALL files including the PHP script
4. ⏳ **Test the contact form** - Make sure PHP is working on your server
5. ⏳ **Consider automated payments** - Gumroad recommended for instant downloads
6. ⏳ **Enable domain whitelisting** in Web3Forms dashboard for security

### Testing Checklist

Before going live:

- [ ] Contact form submits successfully
- [ ] You receive email at info@small-biz-software.com
- [ ] Payment handles are correct
- [ ] All links work (Features, Pricing, Contact navigation)
- [ ] Page looks good on mobile devices
- [ ] No console errors in browser developer tools

## Color Scheme

The page uses a professional, trustworthy palette:
- Primary Blue: #2563eb (trust, stability)
- Dark Blue: #1e40af (professionalism)
- Grays: #f9fafb to #111827 (clean, modern)
- Success Green: #10b981 (positive actions)
- Warning Orange: #f59e0b (important notices)

## Browser Compatibility

- Modern browsers (Chrome, Firefox, Safari, Edge)
- Responsive design (mobile, tablet, desktop)
- Smooth animations (subtle, professional)
- Accessible markup (semantic HTML)
