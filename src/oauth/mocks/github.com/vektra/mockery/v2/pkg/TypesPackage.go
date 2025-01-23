// Code generated by mockery. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// TypesPackage is an autogenerated mock type for the TypesPackage type
type TypesPackage struct {
	mock.Mock
}

type TypesPackage_Expecter struct {
	mock *mock.Mock
}

func (_m *TypesPackage) EXPECT() *TypesPackage_Expecter {
	return &TypesPackage_Expecter{mock: &_m.Mock}
}

// Name provides a mock function with given fields:
func (_m *TypesPackage) Name() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Name")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// TypesPackage_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type TypesPackage_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *TypesPackage_Expecter) Name() *TypesPackage_Name_Call {
	return &TypesPackage_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *TypesPackage_Name_Call) Run(run func()) *TypesPackage_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *TypesPackage_Name_Call) Return(_a0 string) *TypesPackage_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TypesPackage_Name_Call) RunAndReturn(run func() string) *TypesPackage_Name_Call {
	_c.Call.Return(run)
	return _c
}

// Path provides a mock function with given fields:
func (_m *TypesPackage) Path() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Path")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// TypesPackage_Path_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Path'
type TypesPackage_Path_Call struct {
	*mock.Call
}

// Path is a helper method to define mock.On call
func (_e *TypesPackage_Expecter) Path() *TypesPackage_Path_Call {
	return &TypesPackage_Path_Call{Call: _e.mock.On("Path")}
}

func (_c *TypesPackage_Path_Call) Run(run func()) *TypesPackage_Path_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *TypesPackage_Path_Call) Return(_a0 string) *TypesPackage_Path_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TypesPackage_Path_Call) RunAndReturn(run func() string) *TypesPackage_Path_Call {
	_c.Call.Return(run)
	return _c
}

// NewTypesPackage creates a new instance of TypesPackage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTypesPackage(t interface {
	mock.TestingT
	Cleanup(func())
}) *TypesPackage {
	mock := &TypesPackage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
