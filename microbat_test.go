package microbat

import (
	"context"
	"testing"
	"time"
)

func TestMicroBat(t *testing.T) {

	batchProcessor := func(ctx context.Context, job ...Job) ([]JobResult, error) {
		return nil, nil
	}

	client := New(batchProcessor, 2, time.Second*1)
	defer client.Close()

}
