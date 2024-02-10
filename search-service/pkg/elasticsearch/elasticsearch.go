package elasticsearch

import (
	"github.com/JIeeiroSst/search-service/internal"
	esv7 "github.com/elastic/go-elasticsearch/v7"
)

// dns "http://localhost:9200"
func NewElasticSearch(dns string) (es *esv7.Client, err error) {
	es, err = esv7.NewClient(esv7.Config{
		Addresses: []string{dns},
	})
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "elasticsearch.Open")
	}

	res, err := es.Info()
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "es.Info")
	}

	defer func() {
		err = res.Body.Close()
	}()

	return es, nil
}
