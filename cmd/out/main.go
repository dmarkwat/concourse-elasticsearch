package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	concourse "github.com/dmarkwat/concourse-elasticsearch/pkg/concourse"
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

	inputDir := os.Args[1]
	stat, err := os.Stat(inputDir)
	if err != nil {
		log.Fatalf("Error encountered checking argument: %e", err)
	} else if !stat.IsDir() {
		log.Fatalf("%s is not a directory", inputDir)
	}

	request, err := concourse.NewOutRequest(os.Stdin)
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
		log.Printf("Index (%s) doesn't exist; creating...", request.Source.Index)

		err := es.CreateIndex(client, request.Source.Index, request.Params.FieldMap, request.Source.SortFields)
		if err != nil {
			log.Fatal(err)
		}
	}

	fileBytes, err := ioutil.ReadFile(path.Join(inputDir, request.Params.Document))
	if err != nil {
		log.Fatal(err)
	}

	var fileJson map[string]interface{}
	err = json.Unmarshal(fileBytes, &fileJson)
	if err != nil {
		log.Fatal(err)
	}

	digest := sha256.New()
	for _, field := range request.Source.SortFields {
		if _, ok := fileJson[field]; ok {
			log.Fatalf("Sort field, %s, missing from document", field)
		}
		value, ok := fileJson[field].(string)
		if ok {
			log.Fatalf("%s must be a string", field)
		}
		_, err := digest.Write([]byte(value))
		if err != nil {
			log.Fatalf("Error adding to digest: %e", err)
		}
	}

	sum := base64.URLEncoding.EncodeToString(digest.Sum(nil))

	create, err := client.Create(request.Source.Index, sum, bytes.NewReader(fileBytes))
	if err != nil {
		log.Fatalf("Error creating document")
	}
	if create.StatusCode == 409 {
		log.Print("Document already exists; not updating")
		os.Exit(0)
	}

	response := concourse.OutResponse{
		Version: concourse.Version{
			Id: sum,
		},
		Metadata: nil,
	}
	marshal, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}
	if _, err = os.Stdout.Write(marshal); err != nil {
		log.Fatal(err)
	}
}
