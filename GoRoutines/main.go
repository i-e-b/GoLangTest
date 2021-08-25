package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("Let's do things independently...")

	strChannel := make(chan string, 50) // if the capacity is exceeded, senders will block. This can cause deadlocks

	wait := &sync.WaitGroup{}
	wait.Add(2)

	// Not waiting? Go-routines don't hold parent open.
	go printer("A", strChannel, wait)
	go printer("B", strChannel, wait)
	//x := go returner("x") // can't return. Need to get a channel?
	//go returner("x") // can call a returning function, but we lose the value

	// try to pick up a single value from the channel
	if v, ok := <-strChannel; !ok{
		fmt.Println("Failed to get the first value")
	} else{
		fmt.Println("Got the first value:",v)
	}

	//time.Sleep(time.Second) // give the goroutines time to run

	// Get all values until exhausted
	wait.Wait()
	close(strChannel) // `range` over a channel will throw if the channel is not closed
	                  // The message when this happens is "all goroutines are asleep - deadlock!"
	for i := range strChannel {
		fmt.Print(i)
	}


	fmt.Println()
	fmt.Println("Done")
}

func returner(msg string)string{
	return msg+"!"
}

func printer(msg string, channel chan string, wait *sync.WaitGroup) {
	defer wait.Done()
	for i := 0; i < 10; i++ {
		channel <- msg
		fmt.Print(msg)
		time.Sleep(1)
	}
	//close(channel) // if multiple calls try to close one channel, Go will panic
}
