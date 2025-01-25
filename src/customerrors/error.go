package customerrors
import "fmt"

type ErrorDetail struct {
	details map[string]string
}

func NewErrorDetail(code string, description string) *ErrorDetail {
	details := make(map[string]string)
	details["code"] = code
	details["description"] = description
	return &ErrorDetail{
		details: details,
	}
}

func NewErrorDetailWithHttpStatusCode(code string, description string, httpStatusCode int) *ErrorDetail {
	details := make(map[string]string)
	details["code"] = code
	details["description"] = description
	if httpStatusCode >= 100 && httpStatusCode < 600 {
		details["httpStatusCode"] = fmt.Sprint(httpStatusCode)
	}

	return &ErrorDetail{
		details: details,
	}
}
