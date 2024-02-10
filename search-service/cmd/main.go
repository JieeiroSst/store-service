package main

import (
	"context"
	"encoding/json"
	"fmt"

	elastic "github.com/olivere/elastic/v7"
)

func GetESClient() (*elastic.Client, error) {

	client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false))

	fmt.Println("ES initialized...")

	return client, err

}

type Student struct {
	Name         string  `json:"name"`
	Age          int64   `json:"age"`
	AverageScore float64 `json:"average_score"`
}

func main() {
	ctx := context.Background()
	esclient, err := GetESClient()
	if err != nil {
		fmt.Println("Error initializing : ", err)
		panic("Client fail ")
	}
	var students []Student
	searchSource := elastic.NewSearchSource()
	searchSource.Query(elastic.NewMatchQuery("name", "Doe"))
	queryStr, err1 := searchSource.Source()
	queryJs, err2 := json.Marshal(queryStr)
	if err1 != nil || err2 != nil {
		fmt.Println("[esclient][GetResponse]err during query marshal=", err1, err2)
	}
	fmt.Println("[esclient]Final ESQuery=\n", string(queryJs))
	searchService := esclient.Search().Index("students").SearchSource(searchSource)
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		fmt.Println("[ProductsES][GetPIds]Error=", err)
		return
	}
	for _, hit := range searchResult.Hits.Hits {
		var student Student
		err := json.Unmarshal(hit.Source, &student)
		if err != nil {
			fmt.Println("[Getting Students][Unmarshal] Err=", err)
		}

		students = append(students, student)
	}
	if err != nil {
		fmt.Println("Fetching student fail: ", err)
	}

	fmt.Println("=========", students)
}
