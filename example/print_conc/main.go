package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/spid37/microbat"
)

type PrintJob struct {
	id string
}

func (p *PrintJob) ID() string {
	return p.id
}

func (p *PrintJob) Print() {
	fmt.Printf("Processing JOB: %s\n", p.id)
}

type PrintJobResult struct {
	jobID string
}

func (p *PrintJobResult) OK() bool {
	return true
}

func (p *PrintJobResult) JobID() string {
	return p.jobID
}

func main() {
	batchProcessor := func(ctx context.Context, jobs ...microbat.Job) ([]microbat.JobResult, error) {
		var resLock sync.Mutex
		results := []microbat.JobResult{}
		var wg sync.WaitGroup

		for _, job := range jobs {
			switch j := job.(type) {
			case *PrintJob:
				wg.Add(1)
				go func() {
					defer wg.Done()
					j.Print()
					resLock.Lock()
					defer resLock.Unlock()

					results = append(results, &PrintJobResult{jobID: j.id})
				}()
			default:
				return nil, errors.New("unhandled job type")
			}
		}

		wg.Wait()

		return results, nil
	}

	client := microbat.New(batchProcessor, 3, time.Second*1)

	client.AddJob(
		&PrintJob{id: "one"},
		&PrintJob{id: "two"},
		&PrintJob{id: "three"},
		&PrintJob{id: "four"},
		&PrintJob{id: "five"},
		&PrintJob{id: "six"},
	)

	results, err := client.ProcessJobBatches(context.Background())
	if err != nil {
		panic(err)
	}

	println("\nResults:\n")
	for _, res := range results {
		fmt.Printf("result: %s / %t\n", res.JobID(), res.OK())
	}
}
