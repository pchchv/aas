// Code generated by mockery v2.46.3. DO NOT EDIT.

package mocks

import (
	validators "github.com/pchchv/aas/pkg/src/validators"
	mock "github.com/stretchr/testify/mock"
)

// EmailValidator is an autogenerated mock type for the EmailValidator type
type EmailValidator struct {
	mock.Mock
}

// ValidateEmailAddress provides a mock function with given fields: emailAddress
func (_m *EmailValidator) ValidateEmailAddress(emailAddress string) error {
	ret := _m.Called(emailAddress)

	if len(ret) == 0 {
		panic("no return value specified for ValidateEmailAddress")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(emailAddress)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateEmailUpdate provides a mock function with given fields: input
func (_m *EmailValidator) ValidateEmailUpdate(input *validators.ValidateEmailInput) error {
	ret := _m.Called(input)

	if len(ret) == 0 {
		panic("no return value specified for ValidateEmailUpdate")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*validators.ValidateEmailInput) error); ok {
		r0 = rf(input)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewEmailValidator creates a new instance of EmailValidator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEmailValidator(t interface {
	mock.TestingT
	Cleanup(func())
}) *EmailValidator {
	mock := &EmailValidator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
