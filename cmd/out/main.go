package main

import (
	"log"
)

func main() {
	log.Fatal("Not implemented")

	//settings := map[string]interface{}{
	//	"settings": map[string]interface{}{
	//		"index": map[string]interface{}{
	//			"sort.field":   "timestamp",
	//			"sort.order":   "asc",
	//			"sort.missing": "_first",
	//		},
	//	},
	//	"mappings": map[string]interface{}{
	//		"properties": map[string]interface{}{
	//			"timestamp": map[string]interface{}{
	//				"type": "date",
	//			},
	//		},
	//	},
	//}
	//marshal, err := json.Marshal(settings)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//create, err := client.Indices.Create(request.Source.Index, client.Indices.Create.WithBody(bytes.NewReader(marshal)))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//if create.StatusCode != 200 {
	//	log.Fatalf("Error creating index: %s", create.String())
	//}
}
