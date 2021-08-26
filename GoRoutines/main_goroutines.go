package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("Let's do things independently...")

	var value string
	go func() {value = "Ha, ha!"}()
	fmt.Printf("closed value before waiting: '%s'\r\n",value)

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


	// Using an unbuffered channel rather than a waitGroup
	waitChan := make(chan int)
	go pauseAndSend(waitChan)
	fmt.Printf("\r\nWaited for pause (%d)\r\n", <-waitChan)

	// Using a timer ticker
	once := sync.Once{}
	tick := time.Tick(100 * time.Millisecond)
	for i := 0; i < 10; i++ {
		j:= <- tick // feeds date-times
		once.Do(func(){fmt.Printf("Feed of time seconds:")}) // triggers one time per instance of sync.Once
		fmt.Printf(" %v;", j.Second())
	}

	fmt.Println()
	fmt.Printf("closed value after waiting: '%s'\r\n",value)

	fmt.Println("Now for a quick game")
	a := make(chan int) // unbuffered
	b := make(chan int) // unbuffered
	wait.Add(2)
	go PingPonger(a, b, wait, 5)
	go PingPonger(b, a, wait, 15)
	fmt.Printf("S")
	a <- 0// serve. Note, with unbuffered channels, you *MUST* have a listener waiting before you send
	fmt.Printf("*")
	wait.Wait()

	fmt.Println("Done")
}

func PingPonger(inbox chan int, outbox chan int, wait *sync.WaitGroup, times int) {
	fmt.Printf("!")
	defer wait.Done()

	for i := 0; i < times; i++ {
		fmt.Printf("w")
		v,ok := WithTimeout(func()interface{}{return <- inbox}, time.Second)
		j := v.(int)

		if !ok {
			fmt.Printf("... I win! (nw)")
			return
		}

		fmt.Printf("r%d",j)
		_,ok = WithTimeout(func()interface{}{outbox <- j+1; return nil}, time.Second)
		if !ok{
			fmt.Printf("... I win! (nr)")
			return
		}
	}
}

func WithTimeout(delegate func()interface{}, timeout time.Duration) (result interface{}, ok bool) {
	ch := make(chan bool, 1) // buffered
	var ret interface{}
	go func() {
		feedback := make(chan bool) // unbuffered

		// fire delegate on non-waited goroutinue, send 'ok' down the unbuffered channel if it returns
		go func() {
			ret = delegate()
			feedback <- true
		}()

		select { // handle the first of: 1. delegate completes; 2. timeout expires
		case _ = <-feedback: // trigger if got a value
			ch <- true // release the outer wait with success
		case <-time.After(timeout): // trigger if time-out
			ch <- false // release the outer wait with failure
		}
	}()

	ok = <- ch
	return ret, ok
}

func pauseAndSend(waitChan chan int) {
	time.Sleep(time.Second)
	waitChan <- 42
}

func printer(msg string, channel chan string, wait *sync.WaitGroup) {
	//defer func(){wait.Done(); wait.Wait(); close(channel)}() // you can't do this: closing a closed channel causes a panic
	defer wait.Done() // generally always defer, in case we crash somewhere
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