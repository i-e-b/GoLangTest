package playground

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

// Greeting ~is a public mutable (due to `var`) string~ ...it's now a const.
const Greeting string = "Hello, world" // tend toward explicit types on public

func init(){ // This is run when the package gets loaded. It's considered bad practice.
	fmt.Println("Playground was loaded!")
}

type MyDelegate func(int,int)int

func add(a,b int)int{
	return a+b
}

type One struct {}
type two struct {}
type three struct {}
func (receiver One)anna()*two{return new(two) }
func (receiver two)anna()*three{return new(three) }

type Person struct {
	Name string
	Age int
}
type Employee struct {
	Person // Name and Age get directly splatted in ("embedded struct")
	ID string `json:"emp_id" xml:"id" whatever:"x"` // special names for serialisation. These are arbitrary, can have as many as you like, and can reflect them back.
	NextOfKin string `json:"next_of_kin,omitempty"` // don't serialise if empty
	Lookup string `json:",omitempty"` // don't serialise if empty, default name
	PrivateId string `json:"-"` // always omit
}

// https://golang.org/doc/code
func PlaygroundTests() {
	one := new(One)
	fmt.Printf("%T\r\n", one.anna().anna())

	// Composite types
	bloke := Employee{
		Person: Person{ // created with 'deep' syntax. Not optional
			Name: "Tom Noddy",
			Age:  18,
		},
		ID:     "41181G80DY",
	}
	fmt.Println(bloke.Name) // extracted shallow
	fmt.Println(bloke.Person.Name) // extracted shallow
	if str,err:=json.Marshal(bloke); err == nil {
		fmt.Println(string(str)) // serialise
	}
	// Get tag from field
	if field, ok := reflect.TypeOf(bloke)/*.Elem(). <-- if array, slice, pointer, etc*/.FieldByName("ID"); !ok{
		fmt.Println("Couldn't read field from type")
	}else{
		fmt.Printf("%s `%v`", field.Name, field.Tag)
		fmt.Printf(" => %s = %s\r\n", "whatever", field.Tag.Get("whatever"))
	}


	fmt.Println("Sum 1", sum(1, 2, 3, 4))
	fmt.Println("Sum 2", fold(add,1, 2, 3, 4))
	fmt.Println("Product", fold(func(i, j int) int { return i * j }, 1, 2, 3, 4))

	myArray := [3]int{10, 20, 30} //fixed size
	mySlice := []int{1, 2, 3}     // variable size?
	otherSlice := make([]int, 2, 5)
	fmt.Println("Slice before", unsafe.Pointer(&mySlice)) // unsafe.Pointer will allow us see addresses without any protection
	mySlice = append(myArray[:2], mySlice[1:]...) // -> [10 20 2 3]; syntax is [lower-bound : upper-bound] where blank is from start/end. End is exclusive
	fmt.Println("Slice after", unsafe.Pointer(&mySlice))
	//sort.Ints(mySlice)
	fmt.Printf("Slice after mangling: %v;\r\n", mySlice)
	fmt.Printf("Other slice details: len=%d, cap=%d values:%v\r\n", len(otherSlice), cap(otherSlice), otherSlice)
	// NOTE: modifying slices & sub-slices is gross and does weird things
	slice := []int{12,34,56,73}
	subSlice := slice[1:3]
	subSlice = append(subSlice, 42)
	fmt.Println(subSlice, " --> ", slice) // [34 56 42] --> [12 34 56 42] .... it's mad, and runtime dependent
	// Do this instead:
	slice = []int{12,34,56,73}
	subSlice = make([]int,2,5) // if you make this `0,5`, it doesn't work?
	copy(subSlice, /* <--- */ slice[1:3])
	subSlice = append(subSlice, 42)
	fmt.Println(subSlice, " --> ", slice) // [34 56 42]  -->  [12 34 56 73]
	// Slices can pin unused memory if you over-write a large slice with a sub-slice of itself.
	// This will leak if the resulting slice remains in scope

	matrix := [3][4]int{
		{1,2,3,4},
		{4,5,6}, // ends up as {4,5,6,0}; You can't have ragged multiple-dimension ARRAYS - they are rectangular
		{7,8,9,0},
	}
	fmt.Printf("%v\r\n", matrix)

	//mapToSlices := make(map[int][]int, 1)
	// KEY type must be equatable (have `==` and `!=`). VALUE can be anything.
	myMap := map[string]int{"A":1, "B":2} // map[KEY]VALUE{ ... initial values ... }
	fmt.Println(myMap)
	delete(myMap, "A") // remove a key
	myMap["C"] = 3 // set a kvp. VALUE can be anything.
	if v,ok := myMap["Q"]; ok {
		fmt.Println("Q is",v)
	} else {
		fmt.Println("Key 'Q' is not in the map")
	}
	for key,value := range myMap{
		fmt.Printf("%v -> %v; ", key, value)
	}
	fmt.Println()



	//goland:noinspection GoVarAndConstTypeMayBeOmitted
	var n int = 5 // `int` here is optional, inferred from value
	var fibN = Fibonacci(n)
	name := Greeting // same as `var name string = Greeting`

	stuff := "This is something that needs cleaning up"
	defer func() { fmt.Println("I should do this last...", stuff) }() // like `using` in C#
	defer fmt.Println("I should do this last...", stuff)              // but it's callee-based cleanup

	// What happens if you defer in a closure...
	func() {
		defer func() { fmt.Println("I am a defer in a function. What is my scope?") }()
		// the 'defer' above will now fire, as we are ending the scope it was called in.
	}()
	// What happens if you defer in a block...
	if fibN > -100 {
		fmt.Println("Start of a block")
		defer fmt.Println("defer in a block") // this will fire when the outer function ends.
		fmt.Println("End of a block")
	}

	var expandoDenom int = 1
	var expandoNumer float64 = 1.23
	//expandoOutcome := expandoNumer / expandoDenom // no implicit number expansion
	expandoOutcome := expandoNumer / float64(expandoDenom) // but explicit is fine
	fmt.Println("Expanding types -", expandoOutcome)

	var age int
	fmt.Printf("\r\n\r\nAge: ")
	if false {
		_, err := fmt.Scan(&age) // `_` allows us to explicitly ignore a return value. `&` is a pointer ref.
		if err != nil {
			return
		}
	} else {
		age = 22
	}

	if age < 13 {
		fmt.Println("Child")
	} else if age < 20 {
		fmt.Println("Teenager")
	} else if age < 80 {
		fmt.Println("Adult")
	} else {
		fmt.Println("Old")
	}

	//Greeting = "get bent"      <-- not possible, because it's a const
	Greeting := "'Sup mundo?" // <-- shadows the global with a new local variable.
	// looks like Println adds interstitial spaces. Is there a way to turn it off? <-- no.
	bytes, err := fmt.Println(Greeting, fibN, "is the", n, "th fibonacci number.", name, age)
	if err != nil {
		fmt.Println("Wrote", bytes, "bytes. Got error:", err)
	} else {
		fmt.Println("Wrote", bytes, "bytes with no error.")
	}

	_, err2 := Splosion();
	fmt.Printf("Error = %v\r\n", err2)

	//fmt.Println(Greeting, /*main refers to the function, not the namespace*/ main.Greeting, this.Greeting, &Greeting, ...?) // There seems to be no way to escape shadowing
	func(){
		fmt.Println(Greeting) // this gets the shadowed value by closure
	}()
	printGreeting() // this gets the global value due to separate function scope and no closure

	// `Printf` is very much like the C version
	var max = 5
	// cast int to string as `string(...int...)` converts to an utf code point. `Itoa` stringifies the int.
	var maxFibLength = strconv.Itoa( len(fmt.Sprintln(Fibonacci(max)))-1 )
	for i := 1; i <= max; i++ {
		fmt.Printf("    %-"+maxFibLength+"d is the %2d%s fibonaci number\r\n", Fibonacci(i), i, nth(i)) // %2d is pad on left, %-{n}d is pad on right, %s is string
	}

	piGuessStr := "3.141592"
	if piGuess, err := strconv.ParseFloat(piGuessStr, 64); err == nil {
		// `piGuess` and `err` are scoped only to this block
		numErr := math.Abs(math.Pi - piGuess)
		percentErr := (numErr / math.Pi) * 100
		fmt.Printf("The guess was %1.6f%% off\r\n", percentErr) // %f -> float; %{n}.{m}f -> leading zeros, decimal places, %% -> '%'
	}
	// `piGuess` is not accessible here

	/* You can't mix new var defs and existing var references in LHand values. This is pretty annoying
	v,e := strconv.ParseFloat("123.1",64)
	fmt.Println(v,e)
	v2,e := strconv.ParseFloat("321.1",64)
	fmt.Println(v,e)
	*/
	// So you pretty much have to
	var outer float64
	if v,e:= strconv.ParseFloat("123.1",64); e != nil{
		outer += v
	}
	if v,e:= strconv.ParseFloat("321.1",64); e != nil{
		outer += v
	}
	fmt.Println(outer)

	defer calmDown() // any deferred functions will get run after a panic. This lets you clean-up external resources etc.

	// Idiomatic error switch
	switch v,e:= ComplexErrors(age); e { // do the thing, then switch on `e`
	case fishError: fmt.Println("No, that's a fish")
	case creamSauceError: fmt.Println(e.Error())
	case nil:
		fmt.Println(v, e)
	default:
		// Panic is really for where the entire codebase can't continue.
		// Either the code is totally wrong, or we otherwise can' run.
		panic("Unhandled return type") // why don't people do arithmetic types!?
	}

	// the panic recovery still ends this function, it just lets it drop out without killing the app?
	fmt.Println("If you hit the panic age (-255) and can read this, the recover didn't work.")


	rand.Seed(time.Now().UnixNano())
	max  = 100
	min := 10
	v := rand.Intn(max-min) + min
	bullseye :=  rand.Intn(max-min) + min

	fmt.Println(v)
	switch v { // switch with a value is a bit restrictive (but also simpler
	case bullseye:
		fmt.Println("Same again please!")
	case 1,7,12,15:
		fmt.Println("Hey, lucky number")
		/*case v < 20: // doesn't work where we have a `switch [...]`
			fmt.Println("small")
		case v % 2: // this doesn't work to find odd-- just 1 or 0
			fmt.Println("odd?")
			fallthrough*/
	}

	switch { // switch with no values is basically an if chain
	case (v%2)==1:
		fmt.Println("odd")
	default:
		fmt.Println("even")
	}
}

func Lab5_NumberGuessGameError() bool{
	computersGuess := getRandom(1,10)

	showWelcomeMessage()

	success, didCheat := guessLoop(computersGuess, playRound)

	showEndingMessage(success, computersGuess, didCheat)

	return true
}

func DoMagic() string{
	return "Boring reality"
}

func Lab4_NumberGuessGameRefactor() bool{
	computersGuess := getRandom(1,10)

	showWelcomeMessage()

	success, didCheat := guessLoop(computersGuess, playRound)

	showEndingMessage(success, computersGuess, didCheat)

	return true
}

func playRound(computersGuess int, guesses *int) (won bool, usedCheats bool){
	var humanGuess int
	if _, err := fmt.Scan(&humanGuess); err == nil {
		if humanGuess == computersGuess {
			fmt.Println("Well done, you won!")
			fmt.Println("You took", *guesses, plural(*guesses), "to complete the game")
			return true, usedCheats
		} else if humanGuess == -1 {
			*guesses-- // don't count
			fmt.Println("Ok cheater, try ", computersGuess)
			usedCheats = true
		} else {
			fmt.Println("Sorry, wrong number")
			if humanGuess < computersGuess {
				fmt.Println("Your number was lower that the one I want")
			} else {
				fmt.Println("Your number was higher that the one I want")
			}
		}
	} else {
		fmt.Println("Sorry, I didn't understand. Try a number from 1 to 10.")
	}
	return false, usedCheats
}

func showEndingMessage(success bool, computersGuess int, didCheat bool) {
	if !success {
		fmt.Println("Sorry, you ran out of guesses. I was thinking of", computersGuess)
	}
	fmt.Println("Game over")
	if didCheat {
		fmt.Println("You scoundrel")
	}
}

func showWelcomeMessage() {
	fmt.Println("Welcome to the guessing game")
}

func Lab3_NumberGuessGame() bool{
	computersGuess := getRandom(1,10)
	fmt.Println("Welcome to the guessing game")

	success, usedCheats := guessLoop(computersGuess, nil)
	if !success {
		fmt.Println("Sorry, you ran out of guesses. I was thinking of", computersGuess)
	}
	fmt.Println("Game over")
	if usedCheats {fmt.Println("You scoundrel")}

	return true
}

func guessLoop(computersGuess int, round func(computersGuess int, guesses *int) (won bool, usedCheats bool)) (won bool, usedCheats bool) {
	usedCheats = false
	var cheat = false
	for guesses := 1; guesses <= 4; guesses++ {
		fmt.Printf("Please guess a number between 1 and 10: ")

		won, cheat = round(computersGuess, &guesses)
		usedCheats = usedCheats || cheat
		if won {break}
	}
	return false, usedCheats
}

func plural(guesses int) string {
	switch guesses {
	case 1: return "go"
	default:return "goes"
	}
}

func getRandom(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	spread := max-(min-1)
	if spread <= 0 {return 0}
	return rand.Intn(spread) + min
}

func Lab2_Factorial(){
	var baseNum int
	fmt.Println("Starting factorial calculation program")
	fmt.Print("Please input the number: ")
	_, err := fmt.Scan(&baseNum)
	if err != nil {
		fmt.Println("Error!", err)
		return
	}
	if baseNum < 0 || baseNum > 15 {
		fmt.Println("Out of range")
		return
	}

	fmt.Printf("The number to calculate factorial for is %d\n", baseNum)

	var cursor = baseNum-1
	var accum = baseNum
	if accum == 0 {accum = 1}
	for ; cursor > 0; cursor-- {
		accum *= cursor
	}
	fmt.Printf("%d! factorial is %d", baseNum, accum)
}

func nakedReturns() (name1 string, name2 string){
	name1 = "x"
	name2 = "y"

	return
}


func calmDown() {
	if r := recover(); r != nil {
		fmt.Println("recovered from", r)
		//r.???
	}
}

var fishError = errors.New("Fish!")
var creamSauceError = errors.New("Saucey")
var panicError = errors.New("Baaaaaa!")
func ComplexErrors(i int)(int, error){
	if i == 1 {return -1, fishError
	}
	if i == -1 {return 1, creamSauceError
	}
	if i == -255 {return 255, panicError
	}
	return 0, nil
}

func Splosion() (int, error) {
	return 0, errors.New("I don't care, I ain't doing nuffin.")
}

func printGreeting() {
	fmt.Println(Greeting)
}

func fold(f MyDelegate, start int, values ...int) int {
	combined := start
	for _, value := range values {
		combined = f(combined, value)
	}
	return combined
}

func sum(is ...int) (total int) {
	total = 0
	for index, value := range is {
		total += value
		fmt.Printf("%d,",index) // specifically [0..n)
	}
	fmt.Println()
	return
}

func nth(i int) interface{} { // `interface{}` seems to mean 'anything'
	if i == 11 || i == 12 || i == 13 {return "th"}
	var n = i % 10
	if n == 1 {return "st"}
	if n == 2 {return "nd"}
	if n == 3 {return "rd"}
	return "th"
}

func Fibonacci(n int) int {
	var a = 0
	var b = 1
	var c = 0
	for i := 0; i < n; i++ {
		c = a + b
		a = b
		b = c
	}
	return c
}
