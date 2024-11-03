package main

import (
	"context"
	"errors"
	"fmt"
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
		results := []microbat.JobResult{}
		for _, job := range jobs {
			switch j := job.(type) {
			case *PrintJob:
				j.Print()
				results = append(results, &PrintJobResult{jobID: j.id})
			default:
				return nil, errors.New("unhandled job type")
			}
		}

		return results, nil
	}

	client := microbat.New(batchProcessor, 1, time.Second*1)
	defer client.Close()

	client.AddJob(
		&PrintJob{id: "one"},
		&PrintJob{id: "two"},
		&PrintJob{id: "three"},
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
