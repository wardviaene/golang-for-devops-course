package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type mytype struct {
	counter int
	mu      sync.Mutex
}

func main() {
	mytypeInstance := mytype{}
	finished := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(mytypeInstance *mytype) {
			mytypeInstance.mu.Lock()
			fmt.Printf("input counter: %d\n", mytypeInstance.counter)
			mytypeInstance.counter++
			time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
			if mytypeInstance.counter == 5 {
				fmt.Printf("Found counter == 5\n")
			}
			fmt.Printf("output counter: %d\n", mytypeInstance.counter)
			finished <- true
			mytypeInstance.mu.Unlock()
		}(&mytypeInstance)
	}
	for i := 0; i < 10; i++ {
		<-finished
	}
	fmt.Printf("Counter: %d\n", mytypeInstance.counter)
}
