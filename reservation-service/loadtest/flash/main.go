// flash sends n reservation requests for the same last room(s), each from a different user, with c in flight at once,
// and checks that no more of them succeeded than there are rooms.
//
//	flash -url http://reservation-service-svc/api/v1/reservations -n 1000000 -c 3000 -rooms 1 -body '{...}'
//
// It exits 1 if the number of 201s is not exactly -rooms, so it can run as a Kubernetes Job and fail the pipeline.
// Users are the bearer tokens "u<id>", which the stand-in user-service in ../fakeus accepts.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	n := flag.Int("n", 100000, "requests, each from a different user")
	c := flag.Int("c", 2000, "requests in flight at once")
	rooms := flag.Int("rooms", 1, "rooms left: exactly this many requests must succeed")
	url := flag.String("url", "http://127.0.0.1:8080/api/v1/reservations", "endpoint")
	body := flag.String("body", "", "json body")
	keyed := flag.Bool("keyed", false, "send an Idempotency-Key with every request")
	firstUser := flag.Int64("first-user", 100_000, "id of the first user")
	flag.Parse()

	tr := &http.Transport{
		DialContext:         (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		MaxIdleConns:        *c * 2,
		MaxIdleConnsPerHost: *c * 2,
		MaxConnsPerHost:     *c,
		IdleConnTimeout:     30 * time.Second,
	}
	client := &http.Client{Transport: tr, Timeout: 30 * time.Second}

	var next atomic.Int64
	var mu sync.Mutex
	codes := map[int]int{}
	lat := make([]time.Duration, 0, *n)
	var winners []int64
	var errs atomic.Int64

	start := make(chan struct{})
	var wg sync.WaitGroup
	for w := 0; w < *c; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start // every worker is released at the same instant
			for {
				i := next.Add(1) - 1
				if i >= int64(*n) {
					return
				}
				user := *firstUser + i
				req, _ := http.NewRequest("POST", *url, bytes.NewReader([]byte(*body)))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", fmt.Sprintf("Bearer u%d", user))
				if *keyed {
					req.Header.Set("Idempotency-Key", fmt.Sprintf("k-%d", user))
				}
				t0 := time.Now()
				resp, err := client.Do(req)
				d := time.Since(t0)
				if err != nil {
					errs.Add(1)
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				mu.Lock()
				codes[resp.StatusCode]++
				lat = append(lat, d)
				if resp.StatusCode == 201 {
					winners = append(winners, user)
				}
				mu.Unlock()
			}
		}()
	}
	t0 := time.Now()
	close(start)
	wg.Wait()
	total := time.Since(t0)

	sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
	q := func(p float64) time.Duration {
		if len(lat) == 0 {
			return 0
		}
		return lat[int(float64(len(lat)-1)*p)]
	}
	fmt.Printf("requests: %d from %d different users, %d at a time\n", *n, *n, *c)
	fmt.Printf("took %s  =>  %.0f requests/s\n", total.Round(time.Millisecond), float64(*n)/total.Seconds())
	fmt.Printf("status codes: %v   transport errors: %d\n", codes, errs.Load())
	fmt.Printf("winners (201): %v\n", winners)
	fmt.Printf("latency  p50 %s  p90 %s  p99 %s  max %s\n", q(.5).Round(time.Microsecond), q(.9).Round(time.Microsecond), q(.99).Round(time.Microsecond), q(1).Round(time.Microsecond))
	if len(winners) != *rooms {
		fmt.Printf("FAIL: %d reservations succeeded for %d room(s)\n", len(winners), *rooms)
		os.Exit(1)
	}
	fmt.Println("OK: exactly as many reservations as rooms")
}
