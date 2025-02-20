// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/rancher/fleet/pkg/git/poll (interfaces: GitFetcher)
//
// Generated by this command:
//
//	mockgen --build_flags=--mod=mod -destination=../mocks/fetch_mock.go -package=mocks github.com/rancher/fleet/pkg/git/poll GitFetcher
//
// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	v1 "github.com/rancher/fleet/pkg/apis/gitjob.cattle.io/v1"
	gomock "go.uber.org/mock/gomock"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockGitFetcher is a mock of GitFetcher interface.
type MockGitFetcher struct {
	ctrl     *gomock.Controller
	recorder *MockGitFetcherMockRecorder
}

// MockGitFetcherMockRecorder is the mock recorder for MockGitFetcher.
type MockGitFetcherMockRecorder struct {
	mock *MockGitFetcher
}

// NewMockGitFetcher creates a new mock instance.
func NewMockGitFetcher(ctrl *gomock.Controller) *MockGitFetcher {
	mock := &MockGitFetcher{ctrl: ctrl}
	mock.recorder = &MockGitFetcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGitFetcher) EXPECT() *MockGitFetcherMockRecorder {
	return m.recorder
}

// LatestCommit mocks base method.
func (m *MockGitFetcher) LatestCommit(arg0 context.Context, arg1 *v1.GitJob, arg2 client.Client) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LatestCommit", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LatestCommit indicates an expected call of LatestCommit.
func (mr *MockGitFetcherMockRecorder) LatestCommit(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LatestCommit", reflect.TypeOf((*MockGitFetcher)(nil).LatestCommit), arg0, arg1, arg2)
}
