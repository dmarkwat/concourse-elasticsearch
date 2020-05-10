package concourse

import (
	"encoding/json"
	"fmt"
	"io"
)

func NewCheckRequest(reader io.Reader) (*CheckRequest, error) {
	request := CheckRequest{}
	err := json.NewDecoder(reader).Decode(&request)
	if err != nil {
		return nil, err
	}

	if request.Source.Index == "" {
		return nil, fmt.Errorf("invalid source config: index required")
	} else if len(request.Source.Addresses) == 0 {
		return nil, fmt.Errorf("invalid source config: addresses required")
	} else if len(request.Source.SortFields) == 0 {
		return nil, fmt.Errorf("invalid source config: sort_fields required")
	}

	return &request, nil
}
