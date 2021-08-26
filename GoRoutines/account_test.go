package main

import (
	"sync"
	"testing"
	"time"
)

func TestParallelExecutionOnAccount(t *testing.T){
	// new account
	expected :=10_000.01
	myAccount := NewAccount(10_000.01)

	wait := &sync.WaitGroup{}
	doLotsOfDeposits(myAccount, wait)
	doLotsOfWithdrawls(myAccount, wait)
	wait.Wait()

	final := myAccount.GetBalance()
	if final !=expected {
		t.Errorf("Expected %v, but got %v\r\n", expected, final)
	}
}

func doLotsOfDeposits(account *Account, wait *sync.WaitGroup) {
	wait.Add(1)
	go func() {
		for i := 0; i < 1000; i++ {
			account.Deposit(1)
			time.Sleep(time.Millisecond)
		}
		wait.Done()
	}()
}

func doLotsOfWithdrawls(account *Account, wait *sync.WaitGroup) {
	wait.Add(1)
	go func() {
		for i := 0; i < 1000; i++ {
			account.Withdraw(1)
			time.Sleep(time.Millisecond)
		}
		wait.Done()
	}()
}