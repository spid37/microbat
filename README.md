# Microbat

The Microbat library is a simple Go batch processing service. It is designed to receive a list of jobs and process them in batches.

## How to use

To use the Microbat library, you need to define jobs and a batch processor function. 

**Define a Job**

Create a struct to define a job. Jobs must implement the `ID()` string method to match the `microbat.Job` interface.

**Define a JobResult**

Create a struct to define the result of a job. Job results must implement the `JobID()` string and `OK()` bool methods to match the `microbat.JobResult` interface.

**Define a Batch Processor Function**

The batch processor function processes a batch of jobs and returns their results. It must match the microbat.BatchProcessor type.

**Create and Use the Client**

Instantiate the Client with your custom batch processor, add jobs, and process them.

---

### Example.

Print job id text in batches of one with a one second wait inbetween.

```go
import "github.com/spid37/microbat"

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
```

