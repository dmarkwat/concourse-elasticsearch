package es

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"testing"
	"time"
)

func NewTestClient() *elastic.Client {
	es, err := elastic.NewDefaultClient()
	if err != nil {
		return nil
	}
	_, err = es.Cluster.Health(
		es.Cluster.Health.WithWaitForStatus("yellow"),
		es.Cluster.Health.WithTimeout(10*time.Second))
	if err != nil {
		return nil
	}
	return es
}

func NewIndexName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.New().String())
}

func NewIndex(es *elastic.Client, indexPrefix string, settings map[string]interface{}) (string, error) {
	index := NewIndexName(indexPrefix)
	var request []func(*esapi.IndicesCreateRequest)
	if settings != nil {
		body, err := json.Marshal(settings)
		if err != nil {
			return "", err
		}
		request = append(request, es.Indices.Create.WithBody(bytes.NewReader(body)))
	}
	create, err := es.Indices.Create(index, request...)
	if err != nil {
		return "", err
	}
	if create.IsError() {
		return "", errors.New(create.String())
	}
	return index, nil
}

func RefreshIndex(es *elastic.Client, index string) error {
	refresh, err := es.Indices.Refresh(es.Indices.Refresh.WithIndex(index))
	if err != nil {
		return err
	}
	if refresh.IsError() {
		return errors.New(refresh.String())
	}
	return nil
}

func CleanupIndex(t *testing.T, es *elastic.Client, index string) func() {
	return func() {
		res, err := es.Indices.Delete([]string{index})
		if err != nil {
			t.Log(err)
		}
		if res.IsError() {
			t.Log(res.String())
		}
	}
}

func TestIndexExists(t *testing.T) {
	nonemptyIndex := "indexexists"

	es := NewTestClient()

	t.Run("Non-existent index", func(t *testing.T) {
		emptyIndex := "empty"
		exists, err := IndexExists(es, emptyIndex)
		if err != nil {
			t.Fatal(err)
			return
		}
		if exists {
			t.Errorf("Expected %s to not exist", emptyIndex)
			return
		}
	})

	t.Run("Existing index", func(t *testing.T) {
		index, err := NewIndex(es, nonemptyIndex, nil)
		if err != nil {
			t.Error(err)
			return
		}
		t.Cleanup(CleanupIndex(t, es, index))

		exists, err := IndexExists(es, index)
		if err != nil {
			t.Fatal(err)
			return
		}
		if !exists {
			t.Errorf("Expected %s to exist", nonemptyIndex)
			return
		}
	})
}

func TestFindById(t *testing.T) {
	es := NewTestClient()

	t.Run("Document exists", func(t *testing.T) {
		index, err := NewIndex(es, "byid", nil)
		if err != nil {
			t.Fatal(err)
			return
		}
		t.Cleanup(CleanupIndex(t, es, index))

		_, err = es.Create(index, "1", strings.NewReader("{}"))
		if err != nil {
			t.Error(err)
			return
		}

		err = RefreshIndex(es, index)
		if err != nil {
			t.Error(err)
			return
		}

		doc, err := FindById(es, index, "1")
		if err != nil {
			t.Error(err)
			return
		}
		if len(doc) != 0 {
			t.Errorf("Document should be empty")
			return
		}
	})

	t.Run("Document missing", func(t *testing.T) {
		index, err := NewIndex(es, "byid", nil)
		if err != nil {
			t.Fatal(err)
			return
		}
		t.Cleanup(CleanupIndex(t, es, index))

		doc, err := FindById(es, index, "0")
		if err != nil {
			t.Error(err)
			return
		}
		if doc != nil {
			t.Error("Document should not exist")
			return
		}
	})
}

func TestLatestBySortFields(t *testing.T) {
	sortFields := []string{"timestamp"}
	settings := map[string]interface{}{
		"settings": map[string]interface{}{
			"index": map[string]interface{}{
				"sort.field": "timestamp",
				"sort.order": "asc",
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"type":   "date",
					"format": "strict_date_time",
				},
			},
		},
	}
	es := NewTestClient()

	t.Run("All working", func(t *testing.T) {
		index, err := NewIndex(es, "byid", nil)
		if err != nil {
			t.Fatal(err)
			return
		}
		t.Cleanup(CleanupIndex(t, es, index))

		res, err := es.Create(index, "10", strings.NewReader("{\"timestamp\": \"2020-05-10T00:00:00.000Z\"}"))
		if err != nil {
			t.Error(err)
			return
		}
		if res.IsError() {
			t.Error(res.String())
			return
		}
		res, err = es.Create(index, "11", strings.NewReader("{\"timestamp\": \"2020-05-10T01:00:00.000Z\"}"))
		if err != nil {
			t.Error(err)
			return
		}
		if res.IsError() {
			t.Error(res.String())
			return
		}

		err = RefreshIndex(es, index)
		if err != nil {
			t.Error(err)
			return
		}

		doc := map[string]interface{}{
			"timestamp": "2020-05-10T00:00:00.000Z",
		}

		docs, err := LatestBySortFields(es, index, sortFields, doc)
		if err != nil {
			t.Error(err)
			return
		}

		if len(docs) != 2 {
			t.Errorf("Both documents should exist; got %d", len(docs))
			return
		}
	})

	t.Run("Fields missing", func(t *testing.T) {
		index, err := NewIndex(es, "fieldsmissing", settings)
		if err != nil {
			t.Fatal(err)
			return
		}
		t.Cleanup(CleanupIndex(t, es, index))

		res, err := es.Create(index, "9", strings.NewReader("{\"not_the_right_field\": \"2020-05-09T00:00:00.000Z\"}"))
		if err != nil {
			t.Error(err)
			return
		}
		if res.IsError() {
			t.Error(res.String())
			return
		}

		docs, err := LatestBySortFields(es, index, sortFields, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if len(docs) != 0 {
			t.Errorf("No documents should be returned; got %d", len(docs))
			return
		}
	})

	t.Run("No newer documents", func(t *testing.T) {
		index, err := NewIndex(es, "nonewdocs", settings)
		if err != nil {
			t.Fatal(err)
			return
		}
		t.Cleanup(CleanupIndex(t, es, index))

		res, err := es.Create(index, "9", strings.NewReader("{\"timestamp\": \"2020-05-09T00:00:00.000Z\"}"))
		if err != nil {
			t.Error(err)
			return
		}
		if res.IsError() {
			t.Error(res.String())
			return
		}

		res, err = es.Create(index, "10", strings.NewReader("{\"timestamp\": \"2020-05-10T00:00:00.000Z\"}"))
		if err != nil {
			t.Error(err)
			return
		}
		if res.IsError() {
			t.Error(res.String())
			return
		}

		err = RefreshIndex(es, index)
		if err != nil {
			t.Error(err)
			return
		}

		doc := map[string]interface{}{
			"timestamp": "2020-05-10T00:00:00.000Z",
		}

		docs, err := LatestBySortFields(es, index, sortFields, doc)
		if err != nil {
			t.Error(err)
			return
		}
		if len(docs) != 1 {
			t.Errorf("Only 1 document should be returned; got %d", len(docs))
			return
		}
	})

	t.Run("No documents (index/docs deleted)", func(t *testing.T) {
		index, err := NewIndex(es, "nodocs", settings)
		if err != nil {
			t.Error(err)
			return
		}
		t.Cleanup(func() {
			_, err := es.Indices.Delete([]string{index})
			if err != nil {
				t.Log(err)
			}
		})

		err = RefreshIndex(es, index)
		if err != nil {
			t.Error(err)
			return
		}

		docs, err := LatestBySortFields(es, index, sortFields, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if len(docs) != 0 {
			t.Errorf("No documents should be returned; got %d", len(docs))
			return
		}
	})
}

func TestCreateIndex(t *testing.T) {
	fieldMap := map[string]PropertyMapping{
		"timestamp": {
			Type: "date",
		},
	}
	var sortFields []string
	for field, _ := range fieldMap {
		sortFields = append(sortFields, field)
	}
	es := NewTestClient()
	t.Run("Working", func(t *testing.T) {
		name := NewIndexName("testcreateindexworking")
		err := CreateIndex(es, name, fieldMap, sortFields)
		if err != nil {
			t.Error(err)
			return
		}
		t.Cleanup(CleanupIndex(t, es, name))

		ordered := []string{
			`{"timestamp":null}`,
			`{"timestamp":"2020-01-01T00:00:00Z"}`,
			`{"timestamp":"2021-01-01T00:00:00Z"}`,
		}
		for idx, item := range ordered {
			create, err := es.Create(name, strconv.Itoa(idx), bytes.NewReader([]byte(item)))
			if err != nil {
				t.Error(err)
				return
			}
			if create.IsError() {
				t.Errorf("Error creating test document, %d", idx)
				return
			}
		}

		err = RefreshIndex(es, name)
		if err != nil {
			t.Error(err)
			return
		}

		search, err := es.Search(
			es.Search.WithContext(context.Background()),
			es.Search.WithIndex(name),
		)
		if err != nil {
			t.Error(err)
			return
		}
		if search.IsError() {
			t.Errorf("Error searching: %s", search.Status())
			return
		}

		defer search.Body.Close()

		var envelope EnvelopeResponse
		err = json.NewDecoder(search.Body).Decode(&envelope)
		if err != nil {
			t.Error(err)
			return
		}

		for idx, value := range envelope.Hits.Hits {
			idInt, err := strconv.Atoi(value.ID)
			if err != nil {
				t.Error(err)
				return
			}
			if idx != idInt {
				t.Errorf("Document ID, %s, returned out of order", value.ID)
			}
		}
	})
}
