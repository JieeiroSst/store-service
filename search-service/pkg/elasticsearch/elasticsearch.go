package elasticsearch

import (
	"fmt"

	elastic "github.com/olivere/elastic/v7"
)

// dns "http://localhost:9200"
func NewESClient(dns string) (*elastic.Client, error) {
	client, err := elastic.NewClient(elastic.SetURL(dns),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false))

	fmt.Println("ES initialized...")
	return client, err
}
