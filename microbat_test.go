package microbat

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func mockBatchProcessor(ctx context.Context, job ...Job) ([]JobResult, error) {
	return nil, nil
}

func TestMicroBat(t *testing.T) {

	batchProcessor := func(ctx context.Context, job ...Job) ([]JobResult, error) {
		return nil, nil
	}

	client := New(batchProcessor, 2, time.Second*1)
	defer client.Close()

}

type MockJob struct {
	id string
}

func (m *MockJob) ID() string {
	return m.id
}

func TestNextBatch(t *testing.T) {
	client := New(mockBatchProcessor, 2, time.Second*5)

	for i := 0; i < 10; i++ {
		client.AddJob(&MockJob{id: fmt.Sprintf("%d", i)})
	}

	var batchCount int
	for client.JobCount() > 0 {
		jobs := client.GetNextBatch()
		fmt.Printf("got %d jobs\n", len(jobs))
		batchCount++
	}

}
