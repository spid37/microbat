package microbat

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Job is an interface that represents a job to be processed
type Job interface {
	ID() string
}

// JobResult is an interface that represents the result of a job
type JobResult interface {
	JobID() string
	OK() bool
}

// BatchProcessor is a function that processes a batch of jobs
type BatchProcessor func(ctx context.Context, jobs ...Job) ([]JobResult, error)

// New creates a new microbat client
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

// AddJob adds jobs to the queue
func (c *Client) AddJob(jobs ...Job) error {
	c.jobLock.Lock()
	defer c.jobLock.Unlock()

	c.jobs = append(c.jobs, jobs...)
	return nil
}

// JobCount returns the number of jobs in the queue
func (c *Client) JobCount() int {
	c.jobLock.Lock()
	defer c.jobLock.Unlock()

	return len(c.jobs)
}

// GetNextBatch returns the next batch of jobs to be processed
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

// ProcessJob processes a single job
func (c *Client) ProcessJob(ctx context.Context, job Job) (JobResult, error) {
	if job == nil {
		return nil, errors.New("invalid nil job")
	}
	if c.processor == nil {
		return nil, errors.New("no batch processor func defined")
	}

	res, err := c.processor(ctx, job)
	if err != nil {
		return nil, err
	}

	if len(res) != 1 {
		return nil, nil
	}

	return res[0], nil
}

// ProcessJobBatches processes all jobs in the queue
func (c *Client) ProcessJobBatches(ctx context.Context) ([]JobResult, error) {
	if c.processor == nil {
		return nil, errors.New("no batch processor func defined")
	}
	var jobResults []JobResult
	for c.JobCount() > 0 {
		jobs := c.GetNextBatch()
		res, err := c.processor(ctx, jobs...)
		if err != nil {
			return nil, err
		}
		if len(res) > 0 {
			jobResults = append(jobResults, res...)
		}
		if c.waitTime > 0 {
			time.Sleep(c.waitTime)
		}
	}

	return jobResults, nil
}
