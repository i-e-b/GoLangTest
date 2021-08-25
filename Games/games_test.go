package games

import (
	"strings"
	"testing"
)

func ExampleShowWelcomeMessage() {
	ShowWelcomeMessage()
	// Output: Welcome to games
}

func TestUserInput(t *testing.T){
	myReader := strings.NewReader("10\r")

	expected := 10
	input, err := GetUserInput(myReader)

	if err != nil{
		t.Errorf("Expected no error, but got %v",err)
	} else if input != expected {
		t.Errorf("Expected %d, but got %d", expected, input)
	}
}