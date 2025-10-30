package auth

import (
	"strings"
	"testing"
)

type ValidatorTestData struct {
	service *ValidatorService
}

var (
	validatorTestData = ValidatorTestData{
		service: NewValidatorService(),
	}
)

// TestNewValidatorService tests validator service creation
func TestNewValidatorService(t *testing.T) {
	service := NewValidatorService()

	if service == nil {
		t.Fatal("Validator service should not be nil")
	}
}

// TestValidateRegisterInput tests user registration input validation
func TestValidateRegisterInput(t *testing.T) {
	testCases := []struct {
		name        string
		email       string
		password    string
		shouldPass  bool
		description string
	}{
		// Valid cases
		{"valid input", "test@example.com", "password123", true, "valid email and password should pass"},
		{"valid complex email", "user.name+tag@example.co.uk", "mypassword123", true, "complex valid email should pass"},
		{"valid long password", "user@test.com", "verylongpasswordwithmorethan8characters", true, "long password should pass"},

		// Email validation failures
		{"empty email", "", "password123", false, "empty email should fail"},
		{"invalid email no @", "invalid-email", "password123", false, "email without @ should fail"},
		{"invalid email no domain", "user@", "password123", false, "email without domain should fail"},
		{"invalid email no tld", "user@domain", "password123", false, "email without TLD should fail"},
		{"invalid email spaces", "user @example.com", "password123", false, "email with spaces should fail"},
		{"invalid email too long", createLongEmail(), "password123", false, "email longer than 254 chars should fail"},

		// Password validation failures
		{"empty password", "test@example.com", "", false, "empty password should fail"},
		{"short password", "test@example.com", "1234567", false, "password shorter than 8 chars should fail"},
		{"exactly 7 chars", "test@example.com", "1234567", false, "exactly 7 char password should fail"},
		{"exactly 8 chars", "test@example.com", "12345678", true, "exactly 8 char password should pass"},

		// Both invalid
		{"both empty", "", "", false, "both empty should fail"},
		{"both invalid", "invalid-email", "short", false, "both invalid should fail"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatorTestData.service.ValidateRegisterInput(tc.email, tc.password)

			if tc.shouldPass && err != nil {
				t.Errorf("Expected validation to pass for %s, but got error: %v", tc.description, err)
			}

			if !tc.shouldPass && err == nil {
				t.Errorf("Expected validation to fail for %s, but got no error", tc.description)
			}
		})
	}
}

// TestValidateLoginInput tests user login input validation
func TestValidateLoginInput(t *testing.T) {
	testCases := []struct {
		name        string
		email       string
		password    string
		shouldPass  bool
		description string
	}{
		// Valid cases (login is more lenient than register)
		{"valid input", "test@example.com", "password123", true, "valid credentials should pass"},
		{"valid with any email format", "any-email-format", "anypassword", true, "login should accept any email format"},
		{"valid short password", "user@test.com", "123", true, "login should accept any password length"},

		// Invalid cases
		{"empty email", "", "password123", false, "empty email should fail"},
		{"empty password", "test@example.com", "", false, "empty password should fail"},
		{"both empty", "", "", false, "both empty should fail"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatorTestData.service.ValidateLoginInput(tc.email, tc.password)

			if tc.shouldPass && err != nil {
				t.Errorf("Expected login validation to pass for %s, but got error: %v", tc.description, err)
			}

			if !tc.shouldPass && err == nil {
				t.Errorf("Expected login validation to fail for %s, but got no error", tc.description)
			}
		})
	}
}

// TestIsValidEmail tests email validation logic
func TestIsValidEmail(t *testing.T) {
	testCases := []struct {
		email    string
		expected bool
		reason   string
	}{
		// Valid emails
		{"test@example.com", true, "basic email should be valid"},
		{"user.name@example.com", true, "email with dot in username should be valid"},
		{"user+tag@example.com", true, "email with plus sign should be valid"},
		{"user123@example123.com", true, "email with numbers should be valid"},
		{"a@b.co", true, "minimal valid email should work"},
		{"very.long.username@very.long.domain.extension", true, "long but valid email should work"},

		// Invalid emails
		{"", false, "empty email should be invalid"},
		{"plainaddress", false, "email without @ should be invalid"},
		{"@example.com", false, "email without username should be invalid"},
		{"user@", false, "email without domain should be invalid"},
		{"user@domain", false, "email without TLD should be invalid"},
		{"user.@example.com", false, "email ending with dot should be invalid"},
		{"user..name@example.com", false, "email with consecutive dots should be invalid"},
		{"user name@example.com", false, "email with space should be invalid"},
		{"user@domain .com", false, "domain with space should be invalid"},
		{createLongEmail(), false, "email longer than 254 chars should be invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			result := validatorTestData.service.IsValidEmail(tc.email)
			if result != tc.expected {
				t.Errorf("Email '%s': expected %v, got %v (%s)",
					tc.email, tc.expected, result, tc.reason)
			}
		})
	}
}

// TestIsValidPassword tests password validation logic
func TestIsValidPassword(t *testing.T) {
	testCases := []struct {
		password string
		expected bool
		reason   string
	}{
		// Valid passwords
		{"password123", true, "8+ characters should be valid"},
		{"12345678", true, "exactly 8 characters should be valid"},
		{"verylongpassword", true, "long password should be valid"},
		{"pass!@#$%^&*()", true, "password with special chars should be valid"},
		{"пароль123", true, "password with unicode should be valid"},

		// Invalid passwords
		{"", false, "empty password should be invalid"},
		{"1234567", false, "7 characters should be invalid"},
		{"short", false, "short password should be invalid"},
		{"a", false, "single character should be invalid"},
		{"1234567", false, "exactly 7 characters should be invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.password, func(t *testing.T) {
			result := validatorTestData.service.IsValidPassword(tc.password)
			if result != tc.expected {
				t.Errorf("Password '%s': expected %v, got %v (%s)",
					tc.password, tc.expected, result, tc.reason)
			}
		})
	}
}

// TestValidationErrors tests the ValidationErrors struct
func TestValidationErrors(t *testing.T) {
	t.Run("empty validation errors", func(t *testing.T) {
		ve := ValidationErrors{}

		if ve.HasErrors() {
			t.Error("Empty ValidationErrors should not have errors")
		}

		if ve.Error() != "" {
			t.Errorf("Empty ValidationErrors should return empty string, got: %s", ve.Error())
		}
	})

	t.Run("email error only", func(t *testing.T) {
		ve := ValidationErrors{Email: "email is invalid"}

		if !ve.HasErrors() {
			t.Error("ValidationErrors with email error should have errors")
		}

		expected := "email: email is invalid"
		if ve.Error() != expected {
			t.Errorf("Expected error message '%s', got '%s'", expected, ve.Error())
		}
	})

	t.Run("password error only", func(t *testing.T) {
		ve := ValidationErrors{Password: "password too short"}

		if !ve.HasErrors() {
			t.Error("ValidationErrors with password error should have errors")
		}

		expected := "password: password too short"
		if ve.Error() != expected {
			t.Errorf("Expected error message '%s', got '%s'", expected, ve.Error())
		}
	})

	t.Run("both errors", func(t *testing.T) {
		ve := ValidationErrors{
			Email:    "email is invalid",
			Password: "password too short",
		}

		if !ve.HasErrors() {
			t.Error("ValidationErrors with both errors should have errors")
		}

		expected := "email: email is invalid, password: password too short"
		if ve.Error() != expected {
			t.Errorf("Expected error message '%s', got '%s'", expected, ve.Error())
		}
	})
}

// TestValidateRegisterInputErrorMessages tests specific error messages
func TestValidateRegisterInputErrorMessages(t *testing.T) {
	testCases := []struct {
		name             string
		email            string
		password         string
		expectedContains string
	}{
		{"empty email", "", "password123", "email is required"},
		{"invalid email", "invalid-email", "password123", "email format is invalid"},
		{"empty password", "test@example.com", "", "password is required"},
		{"short password", "test@example.com", "short", "password must be at least 8 character long"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatorTestData.service.ValidateRegisterInput(tc.email, tc.password)

			if err == nil {
				t.Errorf("Expected error for %s, but got none", tc.name)
				return
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tc.expectedContains) {
				t.Errorf("Expected error message to contain '%s', got '%s'", tc.expectedContains, errMsg)
			}
		})
	}
}

// TestEmailRegex tests the email regex directly
func TestEmailRegex(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
		"123@456.com",
	}

	invalidEmails := []string{
		"plaintext",
		"@example.com",
		"user@",
		"user@domain",
		"user..name@example.com",
	}

	for _, email := range validEmails {
		if !validatorTestData.service.IsValidEmail(email) {
			t.Errorf("Email regex should match valid email: %s", email)
		}
	}

	for _, email := range invalidEmails {
		if validatorTestData.service.IsValidEmail(email) {
			t.Errorf("Email regex should not match invalid email: %s", email)
		}
	}
}

// Helper functions

// createLongEmail creates an email longer than 254 characters
func createLongEmail() string {
	// Create a very long username part
	longUsername := ""
	for i := 0; i < 250; i++ {
		longUsername += "a"
	}
	return longUsername + "@example.com"
}
