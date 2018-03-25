/*
Copyright 2018 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"vitess.io/vitess/go/stats"
	"vitess.io/vitess/go/vt/servenv"
	"vitess.io/vitess/go/vt/vitessdriver"
)

var (
	queryStats = stats.NewTimings("Queries", "queries 1", "queries 2")
	errorStats = stats.NewCountersWithLabels("Errors", "err 1", "err 2", "err")
)

func main() {
	loadGen()

	port := 24001
	servenv.Port = &port
	fmt.Printf("status on http://localhost:%d/debug/vars\n", port)

	servenv.RunDefault()
}

type load struct {
	query       string
	txTime      time.Duration
	wait        time.Duration
	concurrency int
}

func loadGen() {
	loads := []load{{
		query:       "select * from messages",
		wait:        10 * time.Second,
		concurrency: 3,
	}, {
		query:       "select sleep(1) from dual",
		wait:        10 * time.Second,
		concurrency: 8,
	}, {
		query: "select 1, sleep(0.05) from dual",
		wait:  10 * time.Second,
	}, {
		query: "select sleep(15), 1 from dual",
		wait:  30 * time.Second,
	}, {
		query: "insert into messages(page, time_created_ns, message) values(1, 0, 'bb')",
		wait:  2 * time.Hour,
	}, {
		query: "insert into messages(page, time_created_ns, message) values(2, 0, 'bb')",
		wait:  2 * time.Hour,
	}, {
		query:       "update messages set message=lower('bb') where page=1",
		txTime:      1 * time.Second,
		wait:        10 * time.Second,
		concurrency: 5,
	}, {
		query:       "update messages set message='aa' where page=2",
		txTime:      1 * time.Millisecond,
		wait:        1 * time.Second,
		concurrency: 5,
	}, {
		query:  "update messages set message=upper('cc') where page=1",
		txTime: 10 * time.Second,
		wait:   1 * time.Minute,
	}}
	for _, ld := range loads {
		go oneLoad(ld)
	}
}

func oneLoad(ld load) {
	concurrency := ld.concurrency
	if concurrency == 0 {
		concurrency = 1
	}
	time.Sleep(time.Duration(rand.Intn(int(500 * time.Millisecond))))
	for {
		var wg sync.WaitGroup
		for i := 0; i < concurrency; i++ {
			conn, err := vitessdriver.Open("127.0.0.1:15991", "")
			if err != nil {
				errorStats.Add("conn", 1)
				return
			}

			wg.Add(1)
			go func(conn *sql.DB) {
				defer func() {
					wg.Done()
					conn.Close()
					queryStats.Record(ld.query, time.Now())
				}()
				if strings.HasPrefix(ld.query, "select") {
					if _, err := conn.ExecContext(context.Background(), ld.query, nil); err != nil {
						errorStats.Add("select", 1)
					}
				} else {
					tx, err := conn.BeginTx(context.Background(), nil)
					if err != nil {
						errorStats.Add("begin", 1)
						return
					}
					if _, err := tx.ExecContext(context.Background(), ld.query, nil); err != nil {
						errorStats.Add("dml", 1)
						return
					}
					time.Sleep(ld.txTime)
					if err := tx.Commit(); err != nil {
						errorStats.Add("commit", 1)
					}
				}
			}(conn)
		}
		wg.Wait()
		rnd := rand.Intn(int(500*time.Millisecond)) - int(250*time.Millisecond)
		time.Sleep(ld.wait + time.Duration(rnd))
	}
}
