package mockdata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/JIeeiroSst/search-service/pkg/elasticsearch"
)

type Student struct {
	Name         string  `json:"name"`
	Age          int64   `json:"age"`
	AverageScore float64 `json:"average_score"`
}

func MockDataTest() {
	dns := "http://localhost:9200"
	ctx := context.Background()
	esclient, err := elasticsearch.NewESClient(dns)
	if err != nil {
		fmt.Println("Error initializing : ", err)
		panic("Client fail ")
	}

	//creating student object
	newStudent := Student{
		Name:         "Gopher doe 20",
		Age:          9,
		AverageScore: 88,
	}

	dataJSON, err := json.Marshal(newStudent)
	if err != nil {
		fmt.Println("Error initializing : ", err)
	}

	js := string(dataJSON)
	ind, err := esclient.Index().
		Index("students").
		BodyJson(js).
		Do(ctx)

	if err != nil {
		panic(err)
	}
	fmt.Println("[Elastic][InsertProduct]Insertion Successful", ind)
}
