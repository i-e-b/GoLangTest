package keyvaluestore_test

import (
	keyvaluestore "KeyValueStore"
	"testing"
)

func TestRejectDoubleOpen(t *testing.T){
	keyvaluestore.HardReset()

	if err := keyvaluestore.Open(); err != nil{
		t.Errorf("Opening keystore failed with %v\r\n", err)
	}
	if err := keyvaluestore.Open(); err == nil{
		t.Error("Store should have errored, but did not")
	} else if err != keyvaluestore.StoreAlreadyOpenError{
		t.Errorf("Store should have failed with '%v', but got '%v'\r\n", keyvaluestore.StoreAlreadyOpenError, err)
	}
}

func TestRejectDoubleClose(t *testing.T){
	keyvaluestore.HardReset()

	if err := keyvaluestore.Open(); err != nil{
		t.Errorf("Opening keystore failed with %v\r\n", err)
	}

	if err := keyvaluestore.Close(); err != nil{
		t.Errorf("Closing keystore should have succeeded, but got %v\r\n", err)
	}

	if err := keyvaluestore.Close(); err == nil{
		t.Error("Store should have errored, but did not")
	} else if err != keyvaluestore.StoreNotOpenError{
		t.Errorf("Store should have failed with '%v', but got '%v'\r\n", keyvaluestore.StoreNotOpenError, err)
	}
}

func TestRejectDeleteNonStoredKey(t *testing.T){
	keyvaluestore.HardReset()

	if err := keyvaluestore.Open(); err != nil{
		t.Errorf("Opening keystore failed with %v\r\n", err)
	}

	if err := keyvaluestore.Delete("MissingKey"); err == nil{
		t.Error("Delete should have errored, but did not")
	} else if err != keyvaluestore.KeyNotPresentError{
		t.Errorf("Delete should have failed with '%v', but got '%v'\r\n", keyvaluestore.KeyNotPresentError, err)
	}

	if err := keyvaluestore.Close(); err != nil{
		t.Errorf("Closing keystore should have succeeded, but got %v\r\n", err)
	}
}

func TestOpenSaveReadAndClose(t *testing.T) {
	keyvaluestore.HardReset()

	if err := keyvaluestore.Open(); err != nil{
		t.Errorf("Opening keystore failed with %v\r\n", err)
	}

	if v,err := keyvaluestore.Get("AnyKey"); err == nil{
		t.Errorf("Expected store to be empty, but it returned '%s'", v)
	}

	if err := keyvaluestore.Put("CorrectKey", "CorrectValue"); err != nil {
		t.Errorf("Put failed with %v", err)
	}

	if v,err := keyvaluestore.Get("CorrectKey"); err != nil{
		t.Errorf("Expected store to be empty, but it returned '%s'", v)
	} else if v != "CorrectValue" {
		t.Errorf("Expected 'CorrectValue' but got '%s'", v)
	}

	if err := keyvaluestore.Delete("CorrectKey"); err != nil {
		t.Errorf("Delete failed with %v", err)
	}

	if err := keyvaluestore.Close(); err != nil {
		t.Errorf("Failed to close: %v", err)
	}
}

func TestDoublePutIsOk(t *testing.T){
	keyvaluestore.HardReset()

	if err := keyvaluestore.Open(); err != nil{
		t.Errorf("Opening keystore failed with %v\r\n", err)
	}

	if err := keyvaluestore.Put("CorrectKey", "CorrectValue"); err != nil {
		t.Errorf("Put failed with %v", err)
	}

	if err := keyvaluestore.Put("CorrectKey", "SecondValue"); err != nil {
		t.Errorf("Put failed with %v", err)
	}

	if v,err := keyvaluestore.Get("CorrectKey"); err != nil{
		t.Errorf("Expected store to be empty, but it returned '%s'", v)
	} else if v != "SecondValue" {
		t.Errorf("Expected 'SecondValue' but got '%s'", v)
	}

	if err := keyvaluestore.Close(); err != nil {
		t.Errorf("Failed to close: %v", err)
	}
}