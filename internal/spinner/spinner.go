package spinner

import (
	"fmt"
	"os"
	"time"
)

// Spinner displays an animated spinner to indicate work is in progress.
type Spinner struct {
	frames   []string
	delay    time.Duration
	stopChan chan struct{}
	running  bool
}

// NewSpinner creates a new spinner with a default animation.
func NewSpinner() *Spinner {
	return &Spinner{
		frames:   []string{"|", "/", "-", "\\"},
		delay:    100 * time.Millisecond,
		stopChan: make(chan struct{}),
		running:  false,
	}
}

// Start begins the spinner animation with the given message.
func (s *Spinner) Start(message string) {
	if s.running {
		return
	}

	s.running = true
	go func() {
		frameIndex := 0
		for {
			select {
			case <-s.stopChan:
				// Clear the spinner line
				fmt.Fprintf(os.Stderr, "\r%s \n", message)
				return
			default:
				frame := s.frames[frameIndex%len(s.frames)]
				fmt.Fprintf(os.Stderr, "\r%s %s", frame, message)
				frameIndex++
				time.Sleep(s.delay)
			}
		}
	}()
}

// Stop halts the spinner animation.
func (s *Spinner) Stop() {
	if s.running {
		s.running = false
		s.stopChan <- struct{}{}
	}
}
