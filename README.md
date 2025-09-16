# FreelanceTrackerGo

A Go web application for tracking freelance clients, projects, timesheets, and invoices.

## Quick Start

```bash
# Run the application
go run ./cmd/web

# Access the application
open http://localhost:4000
```

## Email Configuration

FreelanceTrackerGo can send invoices directly to clients via email using Gmail SMTP. Follow these steps to set up email functionality:

### Step 1: Enable Gmail App Passwords

1. **Enable 2-Step Verification on your Gmail account:**
   - Go to [Google Account Settings](https://myaccount.google.com/)
   - Click "Security" in the left sidebar
   - Under "How you sign in to Google", find "2-Step Verification"
   - Click "Get started" and follow the on-screen instructions
   - Click "Turn on" to enable 2-Step Verification

2. **Generate an App Password:**
   - Go directly to [App Passwords](https://myaccount.google.com/apppasswords)
   - You may need to sign in again to verify your identity
   - Enter an app name (e.g., "FreelanceTracker") in the text field
   - Click "Create"
   - **Important**: Copy the 16-character password that appears (no spaces)
   - **Note**: You won't see this password again, so store it securely

**Important Notes:**
- App passwords are only available if 2-Step Verification is enabled
- For work/school accounts, app passwords may be disabled by your administrator
- The 16-character password replaces your regular Gmail password for this application

### Step 2: Configure FreelanceTrackerGo

1. **Start the application:**
   ```bash
   go run ./cmd/web
   ```

2. **Navigate to Settings:**
   - Open http://localhost:4000 in your browser
   - Click "Settings" in the navigation menu
   - Click "Edit Settings"

3. **Configure Email Settings:**
   - Set **Email Enabled** to "Yes"
   - Enter your **SMTP Username**: Your full Gmail address (e.g., `your.email@gmail.com`)
   - Enter your **SMTP Password**: The 16-character app password from Step 1
   - Leave other settings as defaults:
     - **SMTP Host**: `smtp.gmail.com`
     - **SMTP Port**: `587`
     - **From Name**: `FreelanceTracker`
     - **Use TLS**: `Yes`

4. **Save Settings:**
   - Click "Update Settings"
   - You should see a success message

### Step 3: Send Invoice Emails

1. **Ensure clients have email addresses:**
   - Go to "Clients" and edit each client
   - Make sure the "Email" field is filled in

2. **Send an invoice email:**
   - Navigate to a project page
   - In the "Invoices" section, click the ✉ button next to any invoice
   - The email will be sent automatically to the client's email address
   - Success/error messages will appear at the top of the page

### Troubleshooting Email Issues

**"Authentication failed" error:**
- Double-check that 2-Factor Authentication is enabled on your Gmail account
- Verify you're using the app password, not your regular Gmail password
- Make sure the app password was copied correctly (no extra spaces)

**"Client email not found" error:**
- Edit the client and add their email address

**"Failed to connect to SMTP server" error:**
- Check your internet connection
- Verify SMTP settings match the defaults above
- Try generating a new app password

**Email not received by client:**
- Check the client's spam/junk folder
- Verify the client's email address is correct

## Authentication

Authentication is currently disabled. To re-enable:
1. Remove `display:none` property of nav div in main.css
2. Uncomment `IsAuthenticated` in helpers.go.

## Development

### Testing
```bash
go test ./...
```

### Database Migrations
```bash
goose -dir migrations sqlite3 ./freelance_tracker.db up
```

### Code Generation
```bash
sqlc generate
```