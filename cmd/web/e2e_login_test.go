package main

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2EUserLogin tests the complete user login workflow
func TestE2EUserLogin(t *testing.T) {
	ctx := SetupE2ETest(t)

	t.Run("successful login redirects to home page", func(t *testing.T) {
		// Create a test user
		_, err := ctx.CreateTestUser("John Doe", "john@example.com", "password123")
		require.NoError(t, err)

		// Use the Login helper method
		err = ctx.Login("john@example.com", "password123")
		require.NoError(t, err)

		// Should be redirected to home page
		ctx.AssertURL("/")
	})

	t.Run("invalid credentials show error", func(t *testing.T) {
		// Navigate to login page
		ctx.Navigate("/user/login")

		// Try to login with invalid credentials
		err := ctx.FillInput("input[name='email']", "invalid@example.com")
		require.NoError(t, err)

		err = ctx.FillInput("input[name='password']", "wrongpassword")
		require.NoError(t, err)

		// Submit the form
		err = ctx.SubmitForm("form[action='/user/login']")
		require.NoError(t, err)

		// Should stay on login page
		ctx.AssertURL("/user/login")

		// Should see error message
		ctx.AssertElementContainsText("body", "Email or password is incorrect")
	})

	t.Run("logout workflow", func(t *testing.T) {
		// Create and login as a user
		_, err := ctx.CreateTestUser("Jane Doe", "jane@example.com", "password456")
		require.NoError(t, err)

		err = ctx.Login("jane@example.com", "password456")
		require.NoError(t, err)

		// Should be on home page after login
		ctx.AssertURL("/")

		// Navigate to logout - use direct navigation without waiting for load
		// since logout immediately redirects (causing MustWaitLoad to timeout)
		err = rod.Try(func() {
			ctx.Page.Timeout(5 * time.Second).MustNavigate(ctx.ServerURL + "/user/logout")
		})
		require.NoError(t, err)

		// Wait for the redirect chain to complete by polling the URL
		// This is more reliable than a fixed sleep, especially in CI
		foundLogin := false
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			// Use Try to avoid panics from timing issues
			var currentURL string
			err := rod.Try(func() {
				info := ctx.Page.MustInfo()
				currentURL = info.URL
			})
			if err == nil && currentURL == ctx.ServerURL+"/user/login" {
				foundLogin = true
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		require.True(t, foundLogin, "Timeout waiting for redirect to /user/login after logout")

		// Should be redirected to login page after logout
		ctx.AssertURL("/user/login")
	})

	t.Run("empty form fields show validation errors", func(t *testing.T) {
		// Navigate to login page
		ctx.Navigate("/user/login")

		// Submit form without filling fields
		err := ctx.SubmitForm("form[action='/user/login']")
		require.NoError(t, err)

		// Should stay on login page
		ctx.AssertURL("/user/login")

		// Should see validation errors
		body := ctx.Page.MustElement("body").MustText()
		assert.Contains(t, body, "Email is required")
		assert.Contains(t, body, "Password is required")
	})
}

// TestE2EUserSignup tests the user signup workflow
func TestE2EUserSignup(t *testing.T) {
	t.Skip("Signup functionality not implemented in application")
	ctx := SetupE2ETest(t)

	t.Run("successful signup redirects to login", func(t *testing.T) {
		// Navigate to signup page
		ctx.Navigate("/user/signup")

		// Verify we're on the signup page
		ctx.AssertElementExists("form[action='/user/signup']")

		// Fill in signup form
		err := ctx.FillInput("input[name='name']", "New User")
		require.NoError(t, err)

		err = ctx.FillInput("input[name='email']", "newuser@example.com")
		require.NoError(t, err)

		err = ctx.FillInput("input[name='password']", "newpassword123")
		require.NoError(t, err)

		// Submit the form
		err = ctx.SubmitForm("form[action='/user/signup']")
		require.NoError(t, err)

		// Should be redirected to login page
		ctx.AssertURL("/user/login")

		// Should be able to login with new credentials
		err = ctx.Login("newuser@example.com", "newpassword123")
		require.NoError(t, err)

		// Should be on home page after successful login
		ctx.AssertURL("/")
	})

	t.Run("duplicate email shows error", func(t *testing.T) {
		// Create existing user
		_, err := ctx.CreateTestUser("Existing User", "existing@example.com", "password123")
		require.NoError(t, err)

		// Navigate to signup page
		ctx.Navigate("/user/signup")

		// Try to signup with existing email
		err = ctx.FillInput("input[name='name']", "Another User")
		require.NoError(t, err)

		err = ctx.FillInput("input[name='email']", "existing@example.com")
		require.NoError(t, err)

		err = ctx.FillInput("input[name='password']", "password456")
		require.NoError(t, err)

		// Submit the form
		err = ctx.SubmitForm("form[action='/user/signup']")
		require.NoError(t, err)

		// Should stay on signup page
		ctx.AssertURL("/user/signup")

		// Should see error message
		ctx.AssertElementContainsText("body", "Email address is already in use")
	})
}
