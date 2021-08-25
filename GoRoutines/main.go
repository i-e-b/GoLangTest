package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("Let's do things independently...")

	// Make a *buffered* channel.
	var strChannel strChan = make(chan string, 50) // If the capacity is exceeded, senders will block. This can cause deadlocks.
	                                    // If no capacity is set, an un-buffered channel will result, which blocks
	                                    // at both ends (i.e. 1 message at a time) -- this allows simple synchronisation.

	wait := &sync.WaitGroup{}
	wait.Add(2)

	// Not waiting? Go-routines don't hold parent open.
	go printer("A", strChannel, wait)
	go printer("B", strChannel, wait)
	/*
	func returner(msg string)string{
		return msg+"!"
	}*/
	//x := go returner("x") // can't return. Need to get a channel?
	//go returner("x") // can call a returning function, but we lose the value

	// try to pick up a single value from the channel
	if v, ok := <-strChannel; !ok{
		fmt.Println("Failed to get the first value")
	} else{
		fmt.Println("Got the first value:",v)
	}

	// Also, using select/case (inside the try... methods)
	strChannel.trySend(" (From me to myself) ")
	if got,ok := strChannel.tryReceive(); ok{
		fmt.Println("Try: ",got)
	} else {
		fmt.Println("Tried, but failed")
	}

	wait.Wait()
	close(strChannel) // `range` over a channel will throw if the channel is not closed
	                  // The message when this happens is "all goroutines are asleep - deadlock!"
	// Get all values on the buffer
	for i := range strChannel {
		fmt.Print(i)
	}


	fmt.Println()
	fmt.Println("Done")
}

func printer(msg string, channel chan string, wait *sync.WaitGroup) {
	//defer func(){wait.Done(); wait.Wait(); close(channel)}() // you can't do this: closing a closed channel causes a panic
	defer wait.Done()
	for i := 0; i < 10; i++ {
		channel <- msg
		time.Sleep(1)
	}
	//close(channel) // if multiple calls try to close one channel, Go will panic
}

type strChan chan string

func (receiver strChan) trySend(msg string) bool {
	select {
	case receiver <- msg:
		return true
	default:
		return false
	}
}

func (receiver strChan) tryReceive() (msg string, ok bool) {
	select {
	case i := <-receiver:
		return i, true
	default:
		return "", false
	}
}