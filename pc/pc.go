package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Semaphore interface {
	Post()
	Wait()
	GetValue() int
}
type semaphore struct {
	max int
	sem chan bool
}

// Init creates a new semaphore with an initial value and some upper bound.
// This is the only difference compared to POSIX semaphores.
func Init(max, value int) Semaphore {
	s := &semaphore{max: max, sem: make(chan bool, max)}
	for i := 0; i < value; i++ {
		s.Post()
	}
	return s
}

// Post increments the value of the semaphore.
func (s *semaphore) Post() {
	s.sem <- true
}

// Wait decrements the value of the semaphore or blocks until that is possible.
func (s *semaphore) Wait() {
	<-s.sem
}

// GetValue returns the value of the semaphore
func (s *semaphore) GetValue() int {
	return len(s.sem)
}



type buffer struct {
	b                 []int
	size, read, write int
}

func newBuffer(size int) buffer {
	return buffer{
		b:     make([]int, size),
		size:  size,
		read:  0,
		write: 0,
	}
}

func (buffer *buffer) get() int {
	x := buffer.b[buffer.read]
	fmt.Println("Get\t", x, "\t", buffer)
	buffer.read = (buffer.read + 1) % len(buffer.b)
	return x
}

func (buffer *buffer) put(x int) {
	buffer.b[buffer.write] = x
	fmt.Println("Put\t", x, "\t", buffer)
	buffer.write = (buffer.write + 1) % len(buffer.b)
}



func producer2(buffer *buffer, spaceAvailable, workAvailable Semaphore, mutex *sync.Mutex, start, delta int) {
	x := start
	for {
		spaceAvailable.Wait()
		mutex.Lock()
		buffer.put(x)
		x = x + delta


		mutex.Unlock()
		workAvailable.Post()
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	}
}


func consumer2(buffer *buffer, spaceAvailable, workAvailable Semaphore, mutex *sync.Mutex) {
	for {
		workAvailable.Wait()
		mutex.Lock()
		_ = buffer.get()

		spaceAvailable.Post()
		mutex.Unlock()

		time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
	}
}

func main() {
	buffer := newBuffer(5)
	var mutex sync.Mutex

	spaceAvailable := Init(5, 5)
	workAvailable := Init(5,0)

	go producer2(&buffer, spaceAvailable, workAvailable, &mutex, 1, 1)
	go producer2(&buffer, spaceAvailable, workAvailable, &mutex, 1000, -1)

	consumer2(&buffer, spaceAvailable, workAvailable, &mutex)
}
