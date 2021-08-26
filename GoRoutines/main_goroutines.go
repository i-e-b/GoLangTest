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
	go PingPonger("Venus", a, b, wait, 5)
	go PingPonger("Serena", b, a, wait, 15)
	a <- 0// serve. Note, with unbuffered channels, you *MUST* have a listener waiting before you send
	wait.Wait()

	// Fan-out to distribute work
	oneSource := make(chan int, 10)
	for i := 0; i < 10; i++ {oneSource <- i}
	go readChan("One", oneSource)
	go readChan("Two", oneSource)
	go readRangeChan("Three", oneSource)

	// wait for the channel to be empty
	for len(oneSource) > 0 {time.Sleep(time.Millisecond*50)}
	close(oneSource) // can't write to a closed channel, but can `_,ok` style read.
	fmt.Println("...done")


	// Fan-in different sources
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

	// Fan-in then fan-out
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

	// Multi-channel select
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
	fmt.Println("\r\n\r\n")

	// Fan-out then Fan-in
	/*
	          { ---- pool workers ---- }
	               +--> worker1 --+
	               |              |
	gen -[ chFo ]--+--> worker2 --+-->[ chFi ]--> acceptor
	               |              |
	               +--> worker3 --+
	 */
	var poolWorkerId = 100
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
	go processStuff(chFo, chFi, pool.Get().(*workerData))
	go processStuff(chFo, chFi, pool.Get().(*workerData))
	go processStuff(chFo, chFi, pool.Get().(*workerData))

	wait.Add(1)
	go acceptStuff(chFi, wait)
	wait.Wait()

	fmt.Println("All Done")
}

func acceptStuff(input <-chan workStruct, wait *sync.WaitGroup) {
	defer wait.Done()
	for work := range input {
		if work.KillSignal {
			fmt.Printf("acceptor terminating\r\n")
			return
		}

		fmt.Printf("Storing data-- generator-item:%d, worker-item:%d worker-id:%d\r\n", work.GeneratorNum, work.WorkerNum, work.WorkerId)
	}
}

func processStuff(input <-chan workStruct, output chan<- workStruct, workerTempData *workerData) {
	defer func() {output <- workStruct{KillSignal: true}}()

	var i int
	for work := range input {
		if work.KillSignal {
			fmt.Printf("worker %d terminating\r\n", workerTempData.Id)
			return
		}

		time.Sleep(time.Millisecond * 100)
		work.WorkerId = workerTempData.Id
		work.WorkerNum = i
		i++
		output <- work
	}
}

func generateStuff(output chan<- workStruct) {
	for i := 0; i < 10; i++ {
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