package rate_limiter

import (
	"github.com/alitto/pond/v2"
)

// Embeddable is a struct that can be embedded in a struct to provide rate limiting capabilities.
type Embeddable[T any] struct {
	pool pond.ResultPool[T]
}

func NewEmbeddable[T any](maxConcurrentRequests int) *Embeddable[T] {
	return &Embeddable[T]{
		pool: pond.NewResultPool[T](maxConcurrentRequests),
	}
}

func (c *Embeddable[T]) Submit(fn func() T) (T, error) {
	task := c.pool.Submit(fn)
	return task.Wait()
}

func (c *Embeddable[T]) SubmitErr(fn func() (T, error)) (T, error) {
	task := c.pool.SubmitErr(fn)
	return task.Wait()
}

func (c *Embeddable[T]) StopAndWait() {
	c.pool.StopAndWait()
}
