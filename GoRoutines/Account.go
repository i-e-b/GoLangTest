package main

type Account struct {
	balance float64

	depositChannel chan float64
	withdrawChannel chan float64
	balanceChannel chan float64
}

func NewAccount(startingBalance float64)*Account{
	acc := &Account{
		balance:         startingBalance,
		depositChannel:  make(chan float64),
		withdrawChannel: make(chan float64),
		balanceChannel:  make(chan float64),
	}
	go acc.monitor()
	return acc
}

func (acc *Account)Deposit(amount float64){
	acc.depositChannel <- amount
}
func (acc *Account)Withdraw(amount float64){
	acc.withdrawChannel <- amount
}
func (acc *Account)GetBalance() float64{
	return <- acc.balanceChannel
}

func (acc *Account)monitor(){
	for {
		select{
		case amount := <- acc.depositChannel:
			acc.balance += amount
		case amount := <- acc.withdrawChannel:
			acc.balance -= amount
		case acc.balanceChannel <- acc.balance: // this weird pattern is 'in case of listener waiting'
			// not much that can go in here... can log maybe?
		}
	}
}
