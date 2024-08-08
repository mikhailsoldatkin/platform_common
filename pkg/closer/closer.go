package closer

import (
	"log"
	"os"
	"os/signal"
	"sync"
)

var globalCloser = New()

// Add adds functions to the global closer.
func Add(f ...func() error) {
	globalCloser.Add(f...)
}

// Wait waits until all the functions added to the global closer are done.
func Wait() {
	globalCloser.Wait()
}

// CloseAll calls all closer functions added to the global closer.
func CloseAll() {
	globalCloser.CloseAll()
}

// Closer holds functions that need to be called to release resources.
type Closer struct {
	mu    sync.Mutex
	once  sync.Once
	done  chan struct{}
	funcs []func() error
}

// New returns new Closer, if []os.Signal is specified Closer will
// automatically call CloseAll when one of signals is received from OS.
func New(sig ...os.Signal) *Closer {
	c := &Closer{done: make(chan struct{})}
	if len(sig) > 0 {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, sig...)
			<-ch
			signal.Stop(ch)
			c.CloseAll()
		}()
	}
	return c
}

// Add func to closer.
func (c *Closer) Add(f ...func() error) {
	c.mu.Lock()
	c.funcs = append(c.funcs, f...)
	c.mu.Unlock()
}

// Wait blocks until all closer functions are done.
func (c *Closer) Wait() {
	<-c.done
}

// CloseAll calls all functions added to the Closer.
func (c *Closer) CloseAll() {
	// ensure CloseAll is only executed once
	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		// call all Closer funcs async
		errs := make(chan error, len(funcs))
		var wg sync.WaitGroup

		for _, f := range funcs {
			wg.Add(1)
			go func(f func() error) {
				defer wg.Done()
				errs <- f()
			}(f)
		}

		// close the errs channel once all goroutines are done
		go func() {
			wg.Wait()
			close(errs)
		}()

		// receive from the errs channel
		for err := range errs {
			if err != nil {
				log.Println("error returned from Closer:", err)
			}
		}
	})
}
