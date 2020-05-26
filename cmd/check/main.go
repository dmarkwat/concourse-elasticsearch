package main

import (
	"encoding/json"
	"github.com/dmarkwat/concourse-elasticsearch/pkg/concourse"
	"github.com/dmarkwat/concourse-elasticsearch/pkg/es"
	elastic "github.com/elastic/go-elasticsearch/v7"
	"log"
	"os"
)

func indexExists(client *elastic.Client, index string) (bool, error) {
	exists, err := es.IndexExists(client, index)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	return true, nil
}

func getVersions(client *elastic.Client, request *concourse.CheckRequest) ([]concourse.Version, error) {
	var document map[string]interface{}
	if request.Version != nil {
		// initial check
		document, err := es.FindById(client, request.Source.Index, request.Version.Id)
		if err != nil {
			return nil, err
		}

		if document == nil {
			return nil, nil
		}
	}

	ids, err := es.LatestBySortFields(client, request.Source.Index, request.Source.SortFields, document)
	if err != nil {
		return nil, err
	}

	versions := concourse.MapVersion(ids, func(id string) concourse.Version {
		return concourse.Version{
			Id: id,
		}
	})
	return versions, nil
}

func main() {
	request, err := concourse.NewCheckRequest(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	client, err := es.NewClient(request.Source.Addresses, request.Source.Username, request.Source.Password)
	if err != nil {
		log.Fatal(err)
	}

	exists, err := indexExists(client, request.Source.Index)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		log.Println("No versions found; index doesn't exist")
		_, err = os.Stdout.Write([]byte(`[]`))
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	versions, err := getVersions(client, request)
	if err != nil {
		log.Fatal(err)
	}

	if versions == nil {
		_, err = os.Stdout.Write([]byte(`[]`))
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	marshal, err := json.Marshal(versions)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stdout.Write(marshal); err != nil {
		log.Fatal(err)
	}
}
