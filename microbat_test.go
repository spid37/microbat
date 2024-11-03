package microbat

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//
// mocks

type MockJob struct {
	id string
}

func (m *MockJob) ID() string {
	return m.id
}

func newMockJobs(count int) []Job {
	var jobs []Job
	for i := 0; i < count; i++ {
		jobs = append(jobs, &MockJob{id: fmt.Sprintf("%d", i)})
	}

	return jobs
}

func NewMockJobResonseSuccess(jobID string) *MockJobResponseSuccess {
	return &MockJobResponseSuccess{
		jobID: jobID,
	}
}

type MockJobResponseSuccess struct {
	jobID string
}

func (j *MockJobResponseSuccess) JobID() string {
	return j.jobID
}

func (j *MockJobResponseSuccess) OK() bool {
	return true
}

func NewMockJobResonseError(jobID string, err error) *MockJobResponseError {
	return &MockJobResponseError{
		jobID: jobID,
		err:   err,
	}
}

type MockJobResponseError struct {
	jobID string
	err   error
}

func (j *MockJobResponseError) JobID() string {
	return j.jobID
}

func (j *MockJobResponseError) OK() bool {
	return false
}

func (j *MockJobResponseError) Err() error {
	return j.err
}

func mockBatchProcessor(ctx context.Context, jobs ...Job) ([]JobResult, error) {
	results := []JobResult{}
	for _, job := range jobs {
		results = append(results, NewMockJobResonseSuccess(job.ID()))
	}

	return results, nil
}

func mockBatchProcessorWithError(ctx context.Context, jobs ...Job) ([]JobResult, error) {
	results := []JobResult{}
	if len(jobs) == 11 {
		// we dont like to have 11 jobs so we are going to error
		return nil, errors.New("something has gone wrong")
	}
	for _, job := range jobs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			switch job.ID() {
			case "5":
				results = append(results, NewMockJobResonseError(job.ID(), errors.New("error on job 5")))
			case "fail-always":
				return nil, errors.New("an fail-always error")
			default:
				results = append(results, NewMockJobResonseSuccess(job.ID()))
			}
		}
	}

	return results, nil
}

//
// tests

func TestMicroBatClientInit(t *testing.T) {
	client := New(mockBatchProcessor, 2, time.Second*1)
	defer client.Close()
}

func TestNextBatch(t *testing.T) {
	t.Run("five batches of two", func(t *testing.T) {
		client := New(mockBatchProcessor, 2, time.Second*5)
		client.AddJob(newMockJobs(10)...)
		assert.Equal(t, client.JobCount(), 10)

		var batchCount int
		for client.JobCount() > 0 {
			jobs := client.GetNextBatch()
			assert.Len(t, jobs, 2)
			batchCount++
		}
		assert.Equal(t, batchCount, 5)
	})

	t.Run("one batch only when job is less than batch", func(t *testing.T) {
		client := New(mockBatchProcessor, 10, time.Second*5)
		client.AddJob(newMockJobs(5)...)
		assert.Equal(t, client.JobCount(), 5)

		var batchCount int
		for client.JobCount() > 0 {
			jobs := client.GetNextBatch()
			assert.Len(t, jobs, 5)
			batchCount++
		}

		assert.Equal(t, batchCount, 1)
	})

	t.Run("zero batches and zero jobs", func(t *testing.T) {
		client := New(mockBatchProcessor, 10, time.Second*5)

		assert.Equal(t, client.JobCount(), 0)

		jobs := client.GetNextBatch()
		assert.Len(t, jobs, 0)
	})
}

func TestProcessJob(t *testing.T) {
	client := New(mockBatchProcessorWithError, 1, time.Second*3)
	defer client.Close()

	t.Run("process a single job", func(t *testing.T) {
		res, err := client.ProcessJob(context.Background(), &MockJob{id: "1"})
		assert.NoError(t, err)
		assert.Equal(t, res.JobID(), "1")
		assert.True(t, res.OK())
	})

	t.Run("single job fails and return jobresult with error", func(t *testing.T) {
		// id 5 should always fail
		res, err := client.ProcessJob(context.Background(), &MockJob{id: "5"})
		assert.NoError(t, err)
		assert.Equal(t, res.JobID(), "5")
		assert.False(t, res.OK())
		joberr := res.(*MockJobResponseError).Err()
		assert.EqualError(t, joberr, "error on job 5")
	})

	t.Run("single job returns error", func(t *testing.T) {
		client := New(mockBatchProcessorWithError, 1, time.Second*3)
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		// id "fail-always" should always fail with a context error
		res, err := client.ProcessJob(ctx, &MockJob{id: "fail-always"})
		assert.Nil(t, res)
		assert.EqualError(t, err, "an fail-always error")
	})
}

func TestProcessBatchJobs(t *testing.T) {
	t.Run("five successful jobs in batches of two", func(t *testing.T) {
		client := New(mockBatchProcessor, 2, time.Second*0)
		defer client.Close()

		// add 5 jobs
		client.AddJob(newMockJobs(5)...)
		results, err := client.ProcessJobBatches(context.Background())
		assert.NoError(t, err)
		assert.Len(t, results, 5)

		// all jobs should be ok
		for _, res := range results {
			assert.True(t, res.OK())
		}
	})

	t.Run("job processor returns an error with zero results", func(t *testing.T) {
		// 11 job batch should trigger an error
		client := New(mockBatchProcessorWithError, 11, time.Second*0)
		defer client.Close()

		client.AddJob(newMockJobs(11)...)
		results, err := client.ProcessJobBatches(context.Background())
		assert.Error(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("job five should return an error response", func(t *testing.T) {
		client := New(mockBatchProcessorWithError, 1, time.Second*0)
		defer client.Close()

		client.AddJob(newMockJobs(10)...)
		results, err := client.ProcessJobBatches(context.Background())
		assert.NoError(t, err)
		assert.Len(t, results, 10)
		for _, res := range results {
			if res.JobID() == "5" {
				assert.False(t, res.OK())
				err := res.(*MockJobResponseError).Err()
				assert.EqualError(t, err, "error on job 5")
				continue
			}
			assert.True(t, res.OK())
		}
	})

	t.Run("with a context timeout error", func(t *testing.T) {
		client := New(mockBatchProcessorWithError, 1, time.Second*5)
		defer client.Close()

		client.AddJob(newMockJobs(10)...)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		results, err := client.ProcessJobBatches(ctx)
		assert.Error(t, err)
		assert.EqualError(t, err, "context deadline exceeded")
		assert.Len(t, results, 0)
	})
}
