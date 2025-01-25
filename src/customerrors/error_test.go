package customerrors

import (
	"strings"
	"testing"
)

func TestNewErrorDetail(t *testing.T) {
	code := "E001"
	description := "Test error"
	errorDetail := NewErrorDetail(code, description)
	if errorDetail.GetCode() != code {
		t.Errorf("Expected code %s, got %s", code, errorDetail.GetCode())
	}

	if errorDetail.GetDescription() != description {
		t.Errorf("Expected description %s, got %s", description, errorDetail.GetDescription())
	}

	if errorDetail.GetHttpStatusCode() != 0 {
		t.Errorf("Expected HTTP status code 0, got %d", errorDetail.GetHttpStatusCode())
	}
}

func TestNewErrorDetailWithHttpStatusCode(t *testing.T) {
	code := "E002"
	description := "Test error with status code"
	httpStatusCode := 400
	errorDetail := NewErrorDetailWithHttpStatusCode(code, description, httpStatusCode)
	if errorDetail.GetCode() != code {
		t.Errorf("Expected code %s, got %s", code, errorDetail.GetCode())
	}

	if errorDetail.GetDescription() != description {
		t.Errorf("Expected description %s, got %s", description, errorDetail.GetDescription())
	}

	if errorDetail.GetHttpStatusCode() != httpStatusCode {
		t.Errorf("Expected HTTP status code %d, got %d", httpStatusCode, errorDetail.GetHttpStatusCode())
	}
}

func TestNewErrorDetail_LargeValues(t *testing.T) {
	code := strings.Repeat("A", 1000)
	description := strings.Repeat("B", 1000000)
	errorDetail := NewErrorDetail(code, description)
	if errorDetail.GetCode() != code {
		t.Errorf("Expected code of length %d, got length %d", len(code), len(errorDetail.GetCode()))
	}

	if errorDetail.GetDescription() != description {
		t.Errorf("Expected description of length %d, got length %d", len(description), len(errorDetail.GetDescription()))
	}
}

func TestNewErrorDetailWithHttpStatusCode_EdgeCases(t *testing.T) {
	testCases := []struct {
		name           string
		code           string
		description    string
		httpStatusCode int
		expectedCode   int
	}{
		{"Minimum valid HTTP status code", "E006", "Min status", 100, 100},
		{"Maximum valid HTTP status code", "E007", "Max status", 599, 599},
		{"Below minimum HTTP status code", "E008", "Below min", 99, 0},
		{"Above maximum HTTP status code", "E009", "Above max", 600, 0},
		{"Negative HTTP status code", "E010", "Negative", -1, 0},
		{"Zero HTTP status code", "E011", "Zero", 0, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errorDetail := NewErrorDetailWithHttpStatusCode(tc.code, tc.description, tc.httpStatusCode)
			actualCode := errorDetail.GetHttpStatusCode()
			if actualCode != tc.expectedCode {
				t.Errorf("Expected HTTP status code %d, got %d", tc.expectedCode, actualCode)
			}
		})
	}
}
