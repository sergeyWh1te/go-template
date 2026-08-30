package main

import (
	"fmt"
	"sync"
	"time"
)

// bufferSize keeps the producer a little ahead of the consumers.
const bufferSize = 10

// workDelay stands in for however long real work would take.
const workDelay = 500 * time.Millisecond

func producer() chan int {
	ch := make(chan int, bufferSize)

	go func() {
		defer close(ch)

		i := 0
		for {
			ch <- i
			i++
			time.Sleep(workDelay) // simulate some work being done
		}
	}()

	return ch
}

func fanOut(in chan int) (outA, outB chan int) {
	outA = make(chan int)
	outB = make(chan int)

	go func() {
		defer func() {
			close(outA)
			close(outB)
		}()

		for d := range in {
			outA <- d
			outB <- d
		}
	}()

	return outA, outB
}

func main() {
	ch := producer()
	a, b := fanOut(ch)

	var wg sync.WaitGroup

	wg.Go(func() {
		for v := range a {
			println("a:", v)
		}
	})

	wg.Go(func() {
		for v := range b {
			println("b:", v)
		}
	})

	wg.Wait()

	fmt.Println(`Main done`)
}
