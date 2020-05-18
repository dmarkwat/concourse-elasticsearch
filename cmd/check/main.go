package main

import (
	"encoding/json"
	"github.com/dmarkwat/concourse-elasticsearch/pkg/concourse"
	"github.com/dmarkwat/concourse-elasticsearch/pkg/es"
	elastic "github.com/elastic/go-elasticsearch/v7"
	"log"
	"os"
)

func main() {
	request, err := concourse.NewCheckRequest(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	cfg := elastic.Config{
		Addresses: request.Source.Addresses,
		Username:  request.Source.Username,
		Password:  request.Source.Password,
	}

	client, err := elastic.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	info, err := client.Info()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(info)

	exists, err := es.IndexExists(client, request.Source.Index)
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

	document, err := es.FindById(client, request.Source.Index, request.Version.Id)
	if err != nil {
		log.Fatal(err)
	}

	if document == nil {
	}

	ids, err := es.LatestBySortFields(client, request.Source.Index, request.Source.SortFields, document)
	if err != nil {
		log.Fatal(err)
	}

	versions := concourse.MapVersion(ids, func(id string) concourse.Version {
		return concourse.Version{
			Id: id,
		}
	})

	marshal, err := json.Marshal(versions)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stdout.Write(marshal); err != nil {
		log.Fatal(err)
	}
}
