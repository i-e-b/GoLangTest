package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type SauceMessage struct {
	Sender string
	Value int
}

type workStruct struct {
	GeneratorNum int
	WorkerId int
	WorkerNum int
	KillSignal bool
}

type workerData struct {
	Id int
}

func main() {
/*
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
	go PingPonger("Venus", a, b, wait, 5)
	go PingPonger("Serena", b, a, wait, 15)
	a <- 0// serve. Note, with unbuffered channels, you *MUST* have a listener waiting before you send
	wait.Wait()

	//<editor-fold desc="Fan-out to distribute work">

	// ---------------------------
	// Fan-out to distribute work
	// ---------------------------
	oneSource := make(chan int, 10)
	for i := 0; i < 10; i++ {oneSource <- i}
	go readChan("One", oneSource)
	go readChan("Two", oneSource)
	go readRangeChan("Three", oneSource)

	// wait for the channel to be empty
	for len(oneSource) > 0 {time.Sleep(time.Millisecond*50)}
	close(oneSource) // can't write to a closed channel, but can `_,ok` style read.
	fmt.Println("...done")
	//</editor-fold>

	//<editor-fold desc="Fan-in different sources">

	// ---------------------------
	// Fan-in different sources
	// ---------------------------
	oneTarget := make(chan SauceMessage, 10)
	go writeChan("One", oneTarget)
	go writeChan("Two", oneTarget)
	go writeChan("Three", oneTarget)

	time.Sleep(time.Millisecond*50)
	for len(oneTarget) > 0 { // wait until empty
		v:= <-oneTarget
		fmt.Printf("%d from %s; ", v.Value, v.Sender)
	}
	fmt.Println("...done")

	fmt.Println("Closed channel length", len(oneSource), "capacity", cap(oneSource), "ptr", &oneSource)
	fmt.Println("Open channel length", len(oneTarget), "capacity", cap(oneTarget), "ptr", &oneTarget)
	close(oneTarget)
	//</editor-fold>

	//<editor-fold desc="Fan-in then fan-out">


	// ---------------------------
	// Fan-in then fan-out
	// ---------------------------
	grandCentral := make(chan SauceMessage, 20)
	for i := 0; i < 4; i++ {
		wait.Add(1)
		go writeChan2("Prod"+s(i), grandCentral, wait)
	}
	for i := 0; i < 4; i++ {
		go readRangeChan2("Cons"+s(i), grandCentral)
	}
	wait.Wait() // wait for writers to finish

	for len(grandCentral) > 0 { // wait until empty
		time.Sleep(time.Millisecond*10)
	}
	close(grandCentral)
	time.Sleep(time.Millisecond*500) // give readers time to send 'closed' messages
	fmt.Println("...all closed")
	//</editor-fold>

	//<editor-fold desc="Multi-channel select">

	// ---------------------------
	// Multi-channel select
	// ---------------------------
	chan1 := make(chan SauceMessage, 1)
	chan2 := make(chan string, 1)
	chan3 := make(chan int, 1)
	killChan := make(chan bool)
	go readMulti(chan1, chan2, chan3, killChan)
	chan1 <- SauceMessage{Sender: "main", Value: 123}
	chan2 <- "Hello"
	chan3 <- 456

	//close(chan2) // if you close some (but not all) channels here, the `select` in `readMulti` will go nuts. Don't do it!
	time.Sleep(time.Millisecond*100)
	killChan<- true
	fmt.Println()
	//</editor-fold>

	//<editor-fold desc="Pipelines, pattern 1 (fixed)">

	//
	//this func -> [pipeSeg1] --> action1 --> [pipeSeg2] --> action2 --> [pipeSeg3] --> this func
	//
	pipeSeg1 := make(chan int)
	pipeSeg2 := make(chan int)
	pipeSeg3 := make(chan int)
	go action1(pipeSeg1, pipeSeg2) // add 1
	go action2(pipeSeg2, pipeSeg3) // multiply by 2
	pipeSeg1 <- 1
	final := <- pipeSeg3
	fmt.Println("Pipeline produced", final)
	close(pipeSeg1)
	close(pipeSeg2)
	close(pipeSeg3)
	fmt.Println()
	//</editor-fold>

	//<editor-fold desc="Pipelines, pattern 2 (chaining)">

	// ---------------------------
	// Pipelines, pattern 2 (chaining)
	// ---------------------------
	final = <- chainAction2(chainAction1(chainGenerate()))
	fmt.Println("Pipeline/chain produced", final)
	fmt.Println()
	//</editor-fold>

	*/

	//<editor-fold desc="Fan-out then Fan-in">

	wait := &sync.WaitGroup{}
	// ---------------------------
	// Fan-out then Fan-in
	// ---------------------------
	/*
	          { ---- pool workers ---- }
	               +--> worker1 --+
	               |              |
	gen -[ chFo ]--+--> worker2 --+-->[ chFi ]--> acceptor
	               |              |
	               +--> worker3 --+
	 */
	var poolWorkerId = 0
	pool := sync.Pool{
		New: func() interface{} {
			// Pools often contain things like *bytes.Buffer, which are
			// temporary and re-usable.
			poolWorkerId++
			return &workerData{Id: poolWorkerId}
		},
	}

	chFo := make(chan workStruct, 10)
	chFi := make(chan workStruct, 10)
	go generateStuff(chFo)
	for i := 0; i < 4; i++ {
		go passPrimes(chFo, chFi, pool.Get().(*workerData))
	}

	wait.Add(1)
	go acceptStuff(chFi, wait)
	wait.Wait()
	close(chFo)
	close(chFi)
	//</editor-fold>

	fmt.Println("All Done")
}

func chainGenerate() chan int {
	pipeChain := make(chan int, 1) // with this pattern, we have to buffer
	pipeChain <- 1                 // otherwise, this would block
	return pipeChain
}

func chainAction1(input chan int) chan int {
	next := make(chan int)
	go action1(input, next)
	return next
}

func chainAction2(input chan int) chan int {
	next := make(chan int)
	go action2(input, next)
	return next
}

func action1(input chan int, output chan int) {
	v := <-input
	output <- v+1
}

func action2(input chan int, output chan int) {
	v := <-input
	output <- v*2
}

func acceptStuff(input <-chan workStruct, wait *sync.WaitGroup) {
	defer wait.Done()
	for work := range input {
		if work.KillSignal {
			fmt.Printf("acceptor terminating\r\n")
			return
		}

		fmt.Printf("Found prime [ %3d ]            #%-2d by worker %d\r\n", work.GeneratorNum, work.WorkerNum, work.WorkerId)
	}
}

var primes = []int{
	2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,59,
	61,67,71,73,79,83,89,97,101,103,107,109,113,127,
	131,137,139,149,151,157,163,167,173,179,181,191,
	193,197,199,211,223,227,229,233,239,241,251,257,
	263,269,271}
// passPrimes only forwards messages if they're prime (or kill messages)
func passPrimes (input <-chan workStruct, output chan<- workStruct, workerTempData *workerData) {
	defer func() {
		defer calmDown() // just to silence weird thing I'm doing for testing
		output <- workStruct{KillSignal: true}
	}() // cascade the kill signal

	var x int
	for work := range input {
		if work.KillSignal {return}

		// lazy prime check. Forward if it matches
		for i := 0; i < len(primes); i++ {
			if primes[i]== work.GeneratorNum {
				work.WorkerId = workerTempData.Id
				work.WorkerNum = x
				x++
				output <- work
			}
		}

		time.Sleep(time.Millisecond * 100)
	}
}

func generateStuff(output chan<- workStruct) {
	for i := 0; i < 272; i++ {
		output <- workStruct{
			GeneratorNum: i,
			WorkerId:     0,
			WorkerNum:    0,
			KillSignal:   false,
		}
	}

	// just for demo
	time.Sleep(time.Second)
	for i := 0; i < 3; i++ {
		output <- workStruct{KillSignal: true}
	}
}

func calmDown() {
	if r := recover(); r != nil {}
}

func readMulti(chan1 <-chan SauceMessage, chan2 <-chan string, chan3 <-chan int, kill <-chan bool) {
	for {
		select{
		case m1 := <- chan1:
			fmt.Println("Got sauce",m1.Sender,"/",m1.Value)
		case m2:= <- chan2:
			fmt.Println("Got string", m2)
		case m3 := <-chan3:
			fmt.Println("Got int", m3)
		case <- kill: // explicit kill signal
			return
		case <-time.After(time.Millisecond*100):
			fmt.Println("nowt t/o")
		}
	}
}

func writeChan2(name string, source chan<- SauceMessage, wait *sync.WaitGroup) { // write only channel
	defer wait.Done()
	for i := 0; i < 5; i++ {
		source <- SauceMessage{
			Sender: name,
			Value:  i,
		}
		fmt.Printf("%s->%d ", name, i)
		time.Sleep(time.Duration(int(time.Millisecond)*i))
	}
}

func readRangeChan2(name string, source <-chan SauceMessage) {
	for v := range source{
		fmt.Printf("%s (%d,%s) ", name, v.Value, v.Sender)
		time.Sleep(time.Millisecond * 10)
	}
	fmt.Printf("%s c. ", name)
}

func s(i int) string {return strconv.Itoa(i) }

func writeChan(name string, source chan<- SauceMessage) { // write only channel
	for i := 0; i < 3; i++ {
		source <- SauceMessage{
			Sender: name,
			Value:  i,
		}
		time.Sleep(time.Millisecond*10)
	}
}

func readChan(name string, source <-chan int) { // read only channel
	for {
		if v, ok := <-source; !ok {
			fmt.Printf("%s closed\r\n", name)
			return
		} else {
			fmt.Printf("%s (i) got %d; ", name, v)
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func readRangeChan(name string, source <-chan int) { // read only channel
	for v := range source{
		fmt.Printf("%s (r) got %d; ", name, v)
		time.Sleep(time.Millisecond * 10)
	}
	fmt.Printf("%s closed\r\n", name)
}

func PingPonger(name string, inbox <-chan int, outbox chan<- int, wait *sync.WaitGroup, times int) {
	defer wait.Done()
	timeout := time.Millisecond*100

	for i := 0; i < times; i++ {
		v, ok := WithTimeout(func() interface{} { return <-inbox }, timeout)

		j := v.(int) + 1

		if !ok {
			fmt.Printf("... %s wins! (nw)\r\n", name)
			return
		}

		fmt.Printf(" %s ret %d ", name, j)

		_, ok = WithTimeout(func() interface{} { outbox <- j; return nil }, timeout)
		if !ok {
			fmt.Printf("... %s wins! (nr)\r\n", name)
			return
		}
	}
}

func WithTimeout(delegate func() interface{}, timeout time.Duration) (ret interface{}, ok bool) {
	ch := make(chan interface{}, 1) // buffered
	go func() { ch <- delegate() }()
	select {
	case ret = <-ch:
		return ret, true
	case <-time.After(timeout):
	}
	return nil, false
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