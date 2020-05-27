package main

import (
	"encoding/json"
	"github.com/dmarkwat/concourse-elasticsearch/pkg/concourse"
	"github.com/dmarkwat/concourse-elasticsearch/pkg/es"
	elastic "github.com/elastic/go-elasticsearch/v7"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func main() {
	if len(os.Args) != 2 {
		// subtract program name
		log.Fatalf("Expected one argument; got %d", len(os.Args)-1)
	}

	outputDir := os.Args[1]
	stat, err := os.Stat(outputDir)
	if err != nil {
		log.Fatalf("Error encountered checking argument: %e", err)
	} else if !stat.IsDir() {
		log.Fatalf("%s is not a directory", outputDir)
	}

	request, err := concourse.NewInRequest(os.Stdin)
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
		log.Fatalf("Index (%s) doesn't exist", request.Source.Index)
	}

	document, err := es.FindById(client, request.Source.Index, request.Version.Id)
	if err != nil {
		log.Fatal(err)
	}
	if document != nil {
		var outFile string
		if request.Params.Document == "" {
			outFile = path.Join(outputDir, request.Version.Id)
		} else {
			outFile = request.Params.Document
		}

		marshal, err := json.Marshal(document)
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(outFile, marshal, os.FileMode(0400))
		if err != nil {
			log.Fatalf("Error encountered outputting file: %e", err)
		}

		response := concourse.InResponse{
			Version:  request.Version,
			Metadata: nil,
		}
		marshal, err = json.Marshal(response)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := os.Stdout.Write(marshal); err != nil {
			log.Fatal(err)
		}
	} else {
		// missing document
		log.Fatalf("Document (%s) doesn't exist in index (%s)", request.Version.Id, request.Source.Index)
	}
}
