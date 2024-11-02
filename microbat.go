package microbat

import (
	"context"
	"time"
)

type Job interface {
	ID() string
}

type JobResult interface{}

type BatchProcessor func(ctx context.Context, job ...Job) ([]JobResult, error)

func New(processor BatchProcessor, batchSize int, waitTime time.Duration) *Client {
	return &Client{
		processor: processor,
		batchSize: batchSize,
		waitTime:  waitTime,
	}
}

type Client struct {
	processor BatchProcessor
	batchSize int
	waitTime  time.Duration
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) AddJob(job ...Job) error {
	return nil
}
