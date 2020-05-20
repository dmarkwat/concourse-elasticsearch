package concourse

import (
	"io"
	"strings"
	"testing"
)

const badJsonStr = `{\lju`

func badJson(t *testing.T, reader io.Reader, f func(io.Reader) error) {
	t.Run("bad json", func(t *testing.T) {
		err := f(reader)
		if err == nil {
			t.Error("bad json parse should yield error")
		}
	})
}

func TestNewCheckRequest(t *testing.T) {
	badJson(t, strings.NewReader(badJsonStr), func(r io.Reader) error {
		_, err := NewCheckRequest(r)
		return err
	})

	t.Run("Bad validation", func(t *testing.T) {
		_, err := NewCheckRequest(strings.NewReader(`{"source":{"index": "myidx"}}`))
		if err == nil {
			t.Error("Should be missing fields")
			return
		}
		_, err = NewCheckRequest(strings.NewReader(`{"source":{"addresses":["local"]}}`))
		if err == nil {
			t.Error("Should be missing fields")
			return
		}
		_, err = NewCheckRequest(strings.NewReader(`{"source":{"index": "myidx","addresses":["local"]}}`))
		if err == nil {
			t.Error("Should be missing fields")
			return
		}
		_, err = NewCheckRequest(strings.NewReader(`{"source":{"sort_fields":["field"]}}`))
		if err == nil {
			t.Error("Should be missing fields")
			return
		}
	})

	t.Run("Passing", func(t *testing.T) {
		_, err := NewCheckRequest(strings.NewReader(`{"source":{"index": "myidx","addresses":["local"],"sort_fields":["field"]}}`))
		if err != nil {
			t.Error(err)
			return
		}
	})
}

func TestNewInRequest(t *testing.T) {
	badJson(t, strings.NewReader(badJsonStr), func(r io.Reader) error {
		_, err := NewInRequest(r)
		return err
	})
}

func TestNewOutRequest(t *testing.T) {
	badJson(t, strings.NewReader(badJsonStr), func(r io.Reader) error {
		_, err := NewOutRequest(r)
		return err
	})
}
