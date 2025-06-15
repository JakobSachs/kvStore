// The request data-type and its associated utility functions
package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type RequestType uint8

const (
	NoOp  RequestType = 0
	Read  RequestType = 1
	Write RequestType = 2
)

func (t RequestType) marshalJSON() ([]byte, error) {
	var s string
	switch t {
	case NoOp:
		s = "NoOp"
	case Read:
		s = "Read"
	case Write:
		s = "Write"
	default:
		return nil, fmt.Errorf("invalid RequestType: %d", t)
	}
	return []byte(`"` + s + `"`), nil
}

func (t *RequestType) unmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("RequestType should be a string, got %s", data)
	}

	switch s {
	case "NoOp":
		*t = NoOp
	case "Read":
		*t = Read
	case "Write":
		*t = Write
	default:
		return fmt.Errorf("unknown RequestType: %q", s)
	}
	return nil
}

type Request struct {
	Type  RequestType `json:"type"`
	Id    uint64      `json:"id"`
	Key   *string     `json:"key,omitempty"`
	Value *string     `json:"value,omitempty"`
}

func Deserialize(data []byte) (Request, error) {
	reader := bytes.NewReader(data)
	var req Request
	json.NewDecoder(reader).Decode(&req)
	return req, nil
}

func Serialize(req Request) ([]byte, error) {
  return json.Marshal(req)
}
