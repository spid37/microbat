package microbat

import (
	"context"
	"sync"
	"time"
)

type Job interface {
	ID() string
}

type JobResult interface {
	ID() string
}

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
	jobLock   sync.Mutex
	jobs      []Job
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) AddJob(jobs ...Job) error {
	c.jobLock.Lock()
	defer c.jobLock.Unlock()

	c.jobs = append(c.jobs, jobs...)
	return nil
}

func (c *Client) JobCount() int {
	c.jobLock.Lock()
	defer c.jobLock.Unlock()

	return len(c.jobs)
}

func (c *Client) GetNextBatch() []Job {
	c.jobLock.Lock()
	defer c.jobLock.Unlock()

	if len(c.jobs) == 0 {
		return []Job{}
	}

	if len(c.jobs) <= c.batchSize {
		jobs := c.jobs
		c.jobs = []Job{}
		return jobs
	}

	jobs := c.jobs[:c.batchSize]
	c.jobs = c.jobs[c.batchSize:]

	return jobs
}
