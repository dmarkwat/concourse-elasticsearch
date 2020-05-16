package concourse

import (
	"encoding/json"
	"fmt"
	"io"
)

func validateSource(source *SourceConfig) error {
	if source.Index == "" {
		return fmt.Errorf("invalid source config: index required")
	} else if len(source.Addresses) == 0 {
		return fmt.Errorf("invalid source config: addresses required")
	} else if len(source.SortFields) == 0 {
		return fmt.Errorf("invalid source config: sort_fields required")
	}
	return nil
}

func NewCheckRequest(reader io.Reader) (*CheckRequest, error) {
	request := CheckRequest{}
	err := json.NewDecoder(reader).Decode(&request)
	if err != nil {
		return nil, err
	}

	err = validateSource(&request.Source)
	if err != nil {
		return nil, err
	} else {
		return &request, nil
	}
}

func NewInRequest(reader io.Reader) (*InRequest, error) {
	request := InRequest{}
	err := json.NewDecoder(reader).Decode(&request)
	if err != nil {
		return nil, err
	}

	err = validateSource(&request.Source)
	if err != nil {
		return nil, err
	} else {
		return &request, nil
	}
}

func NewOutRequest(reader io.Reader) (*OutRequest, error) {
	request := OutRequest{}
	err := json.NewDecoder(reader).Decode(&request)
	if err != nil {
		return nil, err
	}

	err = validateSource(&request.Source)
	if err != nil {
		return nil, err
	} else {
		return &request, nil
	}
}
