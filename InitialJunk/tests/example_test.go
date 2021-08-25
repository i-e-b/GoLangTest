package tests

import (
	"fmt"
	"goLangTest/playground"
	"strconv"
	"testing"
)

// Example tests & Test functions
// then,
// How to- * skip tests
//         * set-up & clear-down

// Set-up & tear down is done in one lump, with the
// run called explicitly in the middle
func TestMain(m *testing.M){
	// Do setup here
	fmt.Println("Setting up tests...")

	m.Run() // fire off the tests. This is ALL the tests in the WHOLE package

	// Do clean-up here
	fmt.Println("...tearing down tests")
}

/* you can't have more than one TestMain in a package
func TestMainOther(m *testing.M){} // "Wrong test signature
*/

// Example tests expect the stdout to match the comment at the bottom.
// These can be part of the auto documentation system.
// GoLand expects ExampleXxx, where Xxx refers to an available function name.
// These are a bit like test comments in Rust
func ExampleDoMagic() {
	fmt.Println(playground.DoMagic())
	// Output: Boring reality
}

// Skipping a test
func TestNothing(t *testing.T){
	t.Skip("Not yet implemented")
}

// Test functions are more normal xUnit tests
func TestConvert(t *testing.T) {
	t.Log("Hello World")
	actual := playground.DoMagic()
	if actual != "magic" {
		//t.Error("Expected 'magic', but got '#{actual}'")
		t.Errorf("Expected 'magic', but got '%s'", actual)
	}
}

// This seems to be the way to do xUnit 'test-cases'
func TestFibonacci(t *testing.T) {
	var inputs = []int{1, 2, 3, 4, 5, 6, 7, 8}
	var expectedValues = []int{1, 2, 3, 5, 8, 13, 21, 34}

	for idx, inputCase := range inputs {
		expected := expectedValues[idx]
		name := strconv.Itoa(inputCase)
		t.Run("Fib("+name+")", func(t *testing.T) { // if you use the outer `t`- the tests run fine, but the hierarchy doesn't

			if idx == 0 {
				t.Run("Subtest Name", // <-- you can have spaces, but they get converted to underscores
					func(t *testing.T) {
					fmt.Println("I'm in a sub-sub test!")
				})
			}

			actual := playground.Fibonacci(inputCase)
			if actual != expected {
				t.Errorf("Expected %d but got %d for input %d", expected, actual, inputCase)
			}
		})
	}
}
