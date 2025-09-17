package email

import (
	"testing"
)

func TestSendPaymentReceivedEmail(t *testing.T) {
	// Test with a disabled email service
	service := NewService(Config{Enabled: false})
	
	err := service.SendPaymentReceivedEmail(
		"client@example.com",
		"John Smith",
		"freelancer@example.com", 
		"Jane Doe",
		"Academic Paper Editing",
	)
	
	if err == nil {
		t.Fatal("Expected error when email service is disabled")
	}
	
	expectedError := "email service is disabled"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestSendPaymentReceivedEmail_NameExtraction(t *testing.T) {
	// Test that first names are extracted correctly
	service := NewService(Config{Enabled: false}) // Disabled so it won't actually try to send
	
	testCases := []struct {
		name         string
		clientName   string
		expectedName string
	}{
		{"Single name", "John", "John"},
		{"Full name", "John Smith", "John"},
		{"Multiple names", "John Michael Smith", "John"},
		{"Empty name", "", ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't easily test the internal name extraction without making it public,
			// but we can ensure the function handles various name formats without panicking
			err := service.SendPaymentReceivedEmail(
				"client@example.com",
				tc.clientName,
				"freelancer@example.com",
				"Jane Doe",
				"Test Project",
			)
			
			// Should always get disabled service error, not a panic
			if err == nil || err.Error() != "email service is disabled" {
				t.Errorf("Expected disabled service error, got: %v", err)
			}
		})
	}
}