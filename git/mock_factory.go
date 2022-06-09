// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Code generated by MockGen. DO NOT EDIT.
// Source: git/factory.go

// Package git is a generated GoMock package.
package git

import (
	reflect "reflect"

	plumbing "github.com/go-git/go-git/v5/plumbing"
	http "github.com/go-git/go-git/v5/plumbing/transport/http"
	gomock "github.com/golang/mock/gomock"
)

// MockFactory is a mock of Factory interface.
type MockFactory struct {
	ctrl     *gomock.Controller
	recorder *MockFactoryMockRecorder
}

// MockFactoryMockRecorder is the mock recorder for MockFactory.
type MockFactoryMockRecorder struct {
	mock *MockFactory
}

// NewMockFactory creates a new mock instance.
func NewMockFactory(ctrl *gomock.Controller) *MockFactory {
	mock := &MockFactory{ctrl: ctrl}
	mock.recorder = &MockFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFactory) EXPECT() *MockFactoryMockRecorder {
	return m.recorder
}

// GetRepository mocks base method.
func (m *MockFactory) GetRepository(url, path string, reference plumbing.ReferenceName, auth *http.BasicAuth) (plumbing.Hash, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRepository", url, path, reference, auth)
	ret0, _ := ret[0].(plumbing.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRepository indicates an expected call of GetRepository.
func (mr *MockFactoryMockRecorder) GetRepository(url, path, reference, auth interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRepository", reflect.TypeOf((*MockFactory)(nil).GetRepository), url, path, reference, auth)
}