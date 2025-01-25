package customerrors

import "testing"

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
