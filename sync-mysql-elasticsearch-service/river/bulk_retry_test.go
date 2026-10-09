package river

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/JIeeiroSst/sync-mysql-elasticsearch/elastic"
)

func TestDoBulkRetriesTransientItemErrors(t *testing.T) {
	var mu sync.Mutex
	var batches [][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var ids []string
		sc := bufio.NewScanner(req.Body)
		for sc.Scan() {
			line := sc.Text()
			if i := strings.Index(line, `"_id":"`); i >= 0 {
				rest := line[i+7:]
				ids = append(ids, rest[:strings.Index(rest, `"`)])
			}
		}
		io.Copy(io.Discard, req.Body)
		mu.Lock()
		batches = append(batches, ids)
		attempt := len(batches)
		mu.Unlock()

		items := make([]string, 0, len(ids))
		for _, id := range ids {
			switch {
			case id == "2" && attempt == 1:
				items = append(items, fmt.Sprintf(`{"index":{"_index":"t","_id":"%s","status":403,"error":{"type":"cluster_block_exception","reason":"index [t] blocked by: [FORBIDDEN/8/index write (api)];"}}}`, id))
			case id == "3" && attempt == 1:
				items = append(items, fmt.Sprintf(`{"index":{"_index":"t","_id":"%s","status":429,"error":{"type":"es_rejected_execution_exception"}}}`, id))
			case id == "4":
				items = append(items, fmt.Sprintf(`{"index":{"_index":"t","_id":"%s","status":400,"error":{"type":"mapper_parsing_exception"}}}`, id))
			default:
				items = append(items, fmt.Sprintf(`{"index":{"_index":"t","_id":"%s","status":201}}`, id))
			}
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"took":1,"errors":true,"items":[%s]}`, strings.Join(items, ","))
	}))
	defer srv.Close()

	r := &River{es: elastic.NewClient(&elastic.ClientConfig{Addr: strings.TrimPrefix(srv.URL, "http://")})}
	r.ctx, r.cancel = context.WithCancel(context.Background())
	defer r.cancel()

	var reqs []*elastic.BulkRequest
	for _, id := range []string{"1", "2", "3", "4"} {
		reqs = append(reqs, &elastic.BulkRequest{Action: elastic.ActionIndex, Index: "t", Type: "_doc", ID: id, Data: map[string]interface{}{"id": id}})
	}

	if err := r.doBulk(reqs); err != nil {
		t.Fatalf("doBulk: %v", err)
	}
	if len(batches) != 2 {
		t.Fatalf("expected 2 bulk calls, got %d: %v", len(batches), batches)
	}
	if strings.Join(batches[1], ",") != "2,3" {
		t.Errorf("retry batch = %v, want [2 3]", batches[1])
	}
}

func TestDoBulkGivesUpOnPersistentBlock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		io.Copy(io.Discard, req.Body)
		fmt.Fprint(w, `{"took":1,"errors":true,"items":[{"index":{"_index":"t","_id":"1","status":503,"error":{"type":"unavailable_shards_exception"}}}]}`)
	}))
	defer srv.Close()

	r := &River{es: elastic.NewClient(&elastic.ClientConfig{Addr: strings.TrimPrefix(srv.URL, "http://")})}
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.cancel()

	err := r.doBulk([]*elastic.BulkRequest{{Action: elastic.ActionIndex, Index: "t", Type: "_doc", ID: "1", Data: map[string]interface{}{}}})
	if err == nil {
		t.Fatal("expected an error so the river stops without saving the binlog position")
	}
}
