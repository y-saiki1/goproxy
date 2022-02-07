package source

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewVersionNotFoundError(t *testing.T) {
	err := errors.New("error")
	candidates := []struct {
		err         error
		expectedErr *VersionNotFoundError
	}{
		// 0
		{
			err: err,
			expectedErr: &VersionNotFoundError{
				Err: err,
			},
		},
		// 1
		{
			err: nil,
			expectedErr: &VersionNotFoundError{
				Err: nil,
			},
		},
	}

	for i, candidate := range candidates {
		newVersionNotFoundErr := NewVersionNotFoundError(candidate.err)
		assert.Error(t, newVersionNotFoundErr, "candidate %d", i)
		assert.Equal(t, candidate.expectedErr, newVersionNotFoundErr, "candidate %d", i)
	}
}

func Test_VersionNotFoundError_Unwrap(t *testing.T) {
	err := errors.New("error")
	candidates := []struct {
		err         error
		expectedErr error
	}{
		// 0
		{
			err:         err,
			expectedErr: err,
		},
		// 1
		{
			err:         nil,
			expectedErr: nil,
		},
	}

	for i, candidate := range candidates {
		newVersionNotFoundErr := NewVersionNotFoundError(candidate.err)
		unwrappedErr := newVersionNotFoundErr.Unwrap()
		assert.Equal(t, candidate.expectedErr, unwrappedErr, "candidate %d", i)
	}
}

func Test_VersionNotFoundError_Error(t *testing.T) {
	err := errors.New("error")
	candidates := []struct {
		err                  *VersionNotFoundError
		expectedErrorMessage string
	}{
		// 0
		{
			err: &VersionNotFoundError{
				Err: err,
			},
			expectedErrorMessage: "version not found: error",
		},
		// 1
		{
			err: &VersionNotFoundError{
				Err: nil,
			},
			expectedErrorMessage: "version not found",
		},
		// 2
		{
			err:                  nil,
			expectedErrorMessage: "version not found",
		},
	}

	for i, candidate := range candidates {
		errorMessage := candidate.err.Error()
		assert.Equal(t, candidate.expectedErrorMessage, errorMessage, "candidate %d", i)
	}
}

func Test_IsVersionNotFound(t *testing.T) {
	err := errors.New("error")
	candidates := []struct {
		err                       error
		expectedIsVersionNotFound bool
	}{
		// 0
		{
			err: &VersionNotFoundError{
				Err: err,
			},
			expectedIsVersionNotFound: true,
		},
		// 1
		{
			err: &VersionNotFoundError{
				Err: nil,
			},
			expectedIsVersionNotFound: true,
		},
		// 2
		{
			err:                       err,
			expectedIsVersionNotFound: false,
		},
	}

	for i, candidate := range candidates {
		isVersionNotFound := IsVersionNotFound(candidate.err)
		assert.Equal(t, candidate.expectedIsVersionNotFound, isVersionNotFound, "candidate %d", i)
	}
}
