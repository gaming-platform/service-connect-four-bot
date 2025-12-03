package api

import (
	"errors"

	commonv1 "github.com/gaming-platform/api/go/common/v1"
	"google.golang.org/protobuf/proto"
)

type ErrorResponse struct {
	errorResponse commonv1.ErrorResponse
}

func NewErrorResponse(resp []byte) (*ErrorResponse, error) {
	var errResp commonv1.ErrorResponse
	err := proto.Unmarshal(resp, &errResp)
	if err != nil {
		return nil, err
	}

	return &ErrorResponse{errorResponse: errResp}, nil
}

func (e *ErrorResponse) HasViolation(identifier string) bool {
	violations := e.errorResponse.GetViolations()
	for _, v := range violations {
		if v.GetIdentifier() == identifier {
			return true
		}
	}

	return false
}

func (e *ErrorResponse) FirstViolation() string {
	violations := e.errorResponse.GetViolations()
	if len(violations) == 0 {
		return ""
	}

	return violations[0].GetIdentifier()
}

func ErrorResponseToError(resp []byte) error {
	errResp, err := NewErrorResponse(resp)
	if err != nil {
		return err
	}

	return errors.New("violations: " + errResp.FirstViolation())
}
