package es

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	elastic "github.com/elastic/go-elasticsearch/v7"
	"log"
)

func NewClient(addresses []string, username string, password string) (*elastic.Client, error) {
	cfg := elastic.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}

	client, err := elastic.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	info, err := client.Info()
	if err != nil {
		return nil, err
	}
	log.Println(info)
	return client, nil
}

func IndexExists(client *elastic.Client, index string) (bool, error) {
	exists, err := client.Indices.Exists([]string{index})
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}
	// https://www.elastic.co/guide/en/elasticsearch/reference/master/indices-exists.html#indices-exists-api-response-codes
	return exists.StatusCode == 200, nil
}

func FindById(client *elastic.Client, index string, id string) (map[string]interface{}, error) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"_id": id,
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding query: %s", err)
	}
	res, err := client.Search(
		client.Search.WithContext(context.Background()),
		client.Search.WithIndex(index),
		client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting response: %s", err)
	}
	if res.IsError() {
		return nil, errors.New(res.String())
	}
	defer res.Body.Close()

	var envelope EnvelopeResponse
	err = json.NewDecoder(res.Body).Decode(&envelope)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	if envelope.Hits.Total.Value != 1 {
		// it needs to be OK for the document to go missing
		return nil, nil
	}

	var obj map[string]interface{}
	err = json.Unmarshal(envelope.Hits.Hits[0].Source, &obj)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return obj, nil
}

func LatestBySortFields(client *elastic.Client, index string, sortFields []string, document map[string]interface{}) ([]string, error) {
	if len(sortFields) == 0 {
		return nil, fmt.Errorf("must have at least one sorted field")
	}

	var query map[string]interface{}
	if document == nil {
		var sortProcessor []map[string]interface{}
		for _, field := range sortFields {
			sortProcessor = append(sortProcessor, map[string]interface{}{
				field: "desc",
			})
		}

		query = map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
			"sort": sortProcessor,
			"size": 1,
		}
	} else {
		rangeQuery := map[string]interface{}{}
		for _, field := range sortFields {
			value, ok := document[field]
			if !ok {
				return nil, fmt.Errorf("field not found in doc: %s", field)
			}
			rangeQuery[field] = map[string]interface{}{
				"gte": value,
			}
		}

		query = map[string]interface{}{
			"query": map[string]interface{}{
				"range": rangeQuery,
			},
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding query: %s", err)
	}
	log.Printf("Executing query, %s", buf.String())
	res, err := client.Search(
		client.Search.WithContext(context.Background()),
		client.Search.WithIndex(index),
		client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting response: %s", err)
	}
	if res.IsError() {
		return nil, errors.New(res.String())
	}
	defer res.Body.Close()

	var envelope EnvelopeResponse
	err = json.NewDecoder(res.Body).Decode(&envelope)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	var ids []string
	for _, hit := range envelope.Hits.Hits {
		ids = append(ids, hit.ID)
	}

	return ids, nil
}

func CreateIndex(client *elastic.Client, index string, fieldMap map[string]PropertyMapping, sortFields []string) error {
	properties := map[string]interface{}{}
	for key, val := range fieldMap {
		properties[key] = map[string]interface{}{
			"type": val.Type,
		}
	}
	settings := map[string]interface{}{
		"settings": map[string]interface{}{
			"index": map[string]interface{}{
				"sort.field":   sortFields,
				"sort.order":   "asc",
				"sort.missing": "_first",
			},
		},
		"mappings": map[string]interface{}{
			"properties": properties,
		},
	}
	marshal, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	create, err := client.Indices.Create(index, client.Indices.Create.WithBody(bytes.NewReader(marshal)))
	if err != nil {
		return err
	}
	if create.StatusCode != 200 {
		return errors.New(fmt.Sprintf("error creating index: %s", create.String()))
	}
	return nil
}
