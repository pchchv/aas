package customerrors

import (
	"fmt"
	"sort"
	"strings"
)

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

func (e *ErrorDetail) IsError(target *ErrorDetail) bool {
	if target == nil {
		return false
	}

	if len(e.details) != len(target.details) {
		return false
	}

	for key, value := range e.details {
		targetValue, exists := target.details[key]
		if !exists || value != targetValue {
			return false
		}
	}

	return true
}

func (e *ErrorDetail) Error() string {
	if e.details["code"] == "" && e.details["httpStatusCode"] == "" {
		return e.details["description"]
	}

	keys := make([]string, 0, len(e.details))
	for k := range e.details {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var sb strings.Builder
	for _, key := range keys {
		if sb.Len() > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(fmt.Sprintf("%v: %v", key, e.details[key]))
	}

	return sb.String()
}
