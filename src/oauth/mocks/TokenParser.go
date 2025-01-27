// Code generated by mockery v2.46.3. DO NOT EDIT.

package mocks

import (
	oauth "github.com/pchchv/aas/src/oauth"
	mock "github.com/stretchr/testify/mock"

	rsa "crypto/rsa"
)

// TokenParser is an autogenerated mock type for the TokenParser type
type TokenParser struct {
	mock.Mock
}

// DecodeAndValidateTokenString provides a mock function with given fields: token, pubKey, withExpirationCheck
func (_m *TokenParser) DecodeAndValidateTokenString(token string, pubKey *rsa.PublicKey, withExpirationCheck bool) (*oauth.Jwt, error) {
	ret := _m.Called(token, pubKey, withExpirationCheck)

	if len(ret) == 0 {
		panic("no return value specified for DecodeAndValidateTokenString")
	}

	var r0 *oauth.Jwt
	var r1 error
	if rf, ok := ret.Get(0).(func(string, *rsa.PublicKey, bool) (*oauth.Jwt, error)); ok {
		return rf(token, pubKey, withExpirationCheck)
	}
	if rf, ok := ret.Get(0).(func(string, *rsa.PublicKey, bool) *oauth.Jwt); ok {
		r0 = rf(token, pubKey, withExpirationCheck)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth.Jwt)
		}
	}

	if rf, ok := ret.Get(1).(func(string, *rsa.PublicKey, bool) error); ok {
		r1 = rf(token, pubKey, withExpirationCheck)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewTokenParser creates a new instance of TokenParser. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTokenParser(t interface {
	mock.TestingT
	Cleanup(func())
}) *TokenParser {
	mock := &TokenParser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
