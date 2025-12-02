package api

import (
	"errors"

	commonv1 "github.com/gaming-platform/api/go/common/v1"
	"google.golang.org/protobuf/proto"
)

func TransformErrorResponse(resp []byte) error {
	var errResp commonv1.ErrorResponse
	err := proto.Unmarshal(resp, &errResp)
	if err != nil {
		return err
	}

	violations := errResp.GetViolations()
	if len(violations) == 0 {
		return errors.New("no violations found")
	}

	return errors.New(violations[0].GetIdentifier())
}
