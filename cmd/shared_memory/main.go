package main

import (
	"fmt"
	"sync"
	"time"
)

// workDelay stands in for however long real work would take.
const workDelay = 500 * time.Millisecond

type Store struct {
	m    sync.RWMutex
	data string
}

func writer1(s *Store) {
	for {
		time.Sleep(time.Millisecond * 1)
		if s.m.TryLock() {
			s.data += "1:"
			time.Sleep(workDelay) // simulate some work being done
			s.m.Unlock()
		}
	}
}

func writer2(s *Store) {
	for {
		time.Sleep(time.Millisecond * 1)
		if s.m.TryLock() {
			s.data += "2:"
			time.Sleep(workDelay) // simulate some work being done
			s.m.Unlock()
		}
	}
}

func reader(s *Store) {
	for {
		time.Sleep(time.Second)
		s.m.RLock()
		if s.data != "" {
			fmt.Println(s.data)
		}
		s.m.RUnlock()
	}
}

func main() {
	s := &Store{
		m:    sync.RWMutex{},
		data: "",
	}

	var wg = &sync.WaitGroup{}

	wg.Add(3)

	go func() {
		defer wg.Done()

		writer1(s)
	}()

	go func() {
		defer wg.Done()

		writer2(s)
	}()

	go func() {
		defer wg.Done()

		reader(s)
	}()
	wg.Wait()

	fmt.Println(`Main done`)
}
