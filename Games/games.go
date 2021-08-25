package games

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

func ShowWelcomeMessage(){
	fmt.Println("Welcome to games")
}

func PlayNumberGuessingGame() bool{
	computersGuess := getRandom(1,10)

	success, didCheat := guessLoop(computersGuess, playRound)

	showEndingMessage(success, computersGuess, didCheat)

	return true
}

func GetUserInput(src io.Reader) (int, error) {
	var humanGuess int
	_, err := fmt.Fscan(src, &humanGuess)
	return humanGuess, err
}

func playRound(computersGuess int, guesses *int) (won bool, usedCheats bool){
	if humanGuess, err := GetUserInput(os.Stdin); err == nil {
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