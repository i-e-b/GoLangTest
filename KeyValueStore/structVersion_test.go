package keyvaluestore_test

import (
	kvs "KeyValueStore"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestRejectDoubleClose_struct(t *testing.T){
	store := kvs.OpenNew()
	fmt.Println(store.InstanceNum) // just to show public-ness
	//fmt.Println(store.coreMap) // "unexported field"

	if err := kvs.CloseExisting(store); err != nil{
		t.Errorf("Closing keystore should have succeeded, but got %v\r\n", err)
	}

	//if err := kvs.CloseExisting(store); err == nil{
	if err := store.Close(); err == nil{ // using "method" syntax
		t.Error("Store should have errored, but did not")
	} else if err != kvs.StoreNotOpenError{
		t.Errorf("Store should have failed with '%v', but got '%v'\r\n", kvs.StoreNotOpenError, err)
	}
}

func TestRejectDeleteNonStoredKey_struct(t *testing.T){
	store := kvs.OpenNew()

	if err := kvs.DeleteValue(store, "MissingKey"); err == nil{
		t.Error("Delete should have errored, but did not")
	} else if err != kvs.KeyNotPresentError{
		t.Errorf("Delete should have failed with '%v', but got '%v'\r\n", kvs.KeyNotPresentError, err)
	}

	if err := kvs.CloseExisting(store); err != nil{
		t.Errorf("Closing keystore should have succeeded, but got %v\r\n", err)
	}
}

func TestMemberUse(t *testing.T){
	store := kvs.OpenNew()

	// test double open
	if err := store.Open(); err == nil{
		t.Error("Store should have errored, but did not")
	} else if err != kvs.StoreAlreadyOpenError{
		t.Errorf("Store should have failed with '%v', but got '%v'\r\n", kvs.StoreAlreadyOpenError, err)
	}

	// Test the get/put/delete functions
	if v,err := store.Get("AnyKey"); err == nil{
		t.Errorf("Expected store to be empty, but it returned '%s'", v)
	}

	if found:=store.Contains("CorrectKey"); found{
		t.Errorf("Expected key to be missing, but it was found")
	}

	if err := store.Put("CorrectKey", "CorrectValue"); err != nil {
		t.Errorf("Put failed with %v", err)
	}

	if found:=store.Contains("CorrectKey"); !found{
		t.Errorf("Expected key to be found, but it was missing")
	}

	fmt.Println(store) // test "Stringer" interface implementation

	if v,err := store.Get("CorrectKey"); err != nil{
		t.Errorf("Expected store to be empty, but it returned '%s'", v)
	} else if v != "CorrectValue" {
		t.Errorf("Expected 'CorrectValue' but got '%s'", v)
	}

	if err := store.Delete("CorrectKey"); err != nil {
		t.Errorf("Delete failed with %v", err)
	}


	// Close, and test double close
	if err := store.Close(); err != nil{
		t.Errorf("Closing keystore should have succeeded, but got %v\r\n", err)
	}

	if err := store.Close(); err == nil{ // using "method" syntax
		t.Error("Store should have errored, but did not")
	} else if err != kvs.StoreNotOpenError{
		t.Errorf("Store should have failed with '%v', but got '%v'\r\n", kvs.StoreNotOpenError, err)
	}
}

func TestOpenSaveReadAndClose_struct(t *testing.T) {
	store := kvs.OpenNew()

	if v,err := kvs.GetValue(store, "AnyKey"); err == nil{
		t.Errorf("Expected store to be empty, but it returned '%s'", v)
	}

	if err := kvs.PutValue(store, "CorrectKey", "CorrectValue"); err != nil {
		t.Errorf("Put failed with %v", err)
	}

	fmt.Println(store) // test "Stringer" interface implementation

	if v,err := kvs.GetValue(store, "CorrectKey"); err != nil{
		t.Errorf("Expected store to be empty, but it returned '%s'", v)
	} else if v != "CorrectValue" {
		t.Errorf("Expected 'CorrectValue' but got '%s'", v)
	}

	if err := kvs.DeleteValue(store, "CorrectKey"); err != nil {
		t.Errorf("Delete failed with %v", err)
	}

	if err := kvs.CloseExisting(store); err != nil {
		t.Errorf("Failed to close: %v", err)
	}
}

func TestDoublePutIsOk_strut(t *testing.T){
	store := kvs.OpenNew()

	if err := kvs.PutValue(store, "CorrectKey", "CorrectValue"); err != nil {
		t.Errorf("Put failed with %v", err)
	}

	if err := kvs.PutValue(store, "CorrectKey", "SecondValue"); err != nil {
		t.Errorf("Put failed with %v", err)
	}

	if v,err := kvs.GetValue(store, "CorrectKey"); err != nil{
		t.Errorf("Expected store to be empty, but it returned '%s'", v)
	} else if v != "SecondValue" {
		t.Errorf("Expected 'SecondValue' but got '%s'", v)
	}

	if err := kvs.CloseExisting(store); err != nil {
		t.Errorf("Failed to close: %v", err)
	}
}

func TestTwoStoresDontInteract_strut(t *testing.T){
	store1 := kvs.OpenNew()
	store2 := kvs.OpenNew()

	// Put in 1
	if err := kvs.PutValue(store1, "CorrectKey", "CorrectValue"); err != nil {
		t.Errorf("Put failed with %v", err)
	}

	// check 1
	if v,err := kvs.GetValue(store1, "CorrectKey"); err != nil{
		t.Errorf("Expected store to be empty, but it returned '%s'", v)
	} else if v != "CorrectValue" {
		t.Errorf("Expected 'CorrectValue' but got '%s'", v)
	}

	// check 2 (should be missing)
	if v,err := kvs.GetValue(store2, "CorrectKey"); err == nil{
		t.Errorf("Get should have errored, but did not")
	} else if err != kvs.KeyNotPresentError{
		t.Errorf("Expected '%s' but got '%s'", kvs.KeyNotPresentError, v)
	}

	// close both
	if err := kvs.CloseExisting(store1); err != nil {
		t.Errorf("Failed to close: %v", err)
	}

	if err := kvs.CloseExisting(store2); err != nil {
		t.Errorf("Failed to close: %v", err)
	}
}

func TestManyKeys(t *testing.T){
	store := kvs.OpenNew()

	for i := 0; i < 1000; i++ {
		if err:= store.Put(kvs.StoreKey("key"+s(i)), "value"+s(i)); err != nil {
			t.Errorf("Put %d failed with %v", i, err)
		}
	}
}

func TestEvictionAndTiming(t *testing.T){
	store := kvs.OpenNew()

	fmt.Println(kvs.Describe(kvs.StoreKey("keep-key")))

	// Put in a range of old and new keys
	if err:= store.Put(kvs.StoreKey("keep-key"), "value"); err != nil {
		t.Errorf("Put failed with %v", err)
	}

	if err:= store.PutWithAge(kvs.StoreKey("refreshed-key"), "value", time.Now().Add(time.Minute)); err != nil {
		t.Errorf("Put failed with %v", err)
	}

	if err:= store.PutWithAge(kvs.StoreKey("lose-key"), "value", time.Now().Add(time.Minute)); err != nil {
		t.Errorf("Put failed with %v", err)
	}

	// 'Get' one of the older keys (it should be 'new' again)
	if v,err := store.Get("refreshed-key"); err != nil{
		t.Errorf("Expected value, but got error '%v'", err)
	} else {
		fmt.Println(kvs.Describe(v))
		v, _ := store.GetAge("refreshed-key")
		fmt.Printf("Latest time for refreshed key = %v\r\n", v)
		// Note: if you forget to put a new-line on the last printf in a test, GoLand won't understand the result
		//       and will show the test as result as " (-) Terminated "
	}

	// Now evict 'old' keys
	store.EvictOlderThan(time.Now().Add(time.Second * 30))


	// Check we have the expected keys
	if found := store.Contains("lose-key"); found {t.Errorf("Expected 'lose-key' to be missing, but it was found")}
	if found := store.Contains("refreshed-key"); !found {t.Errorf("Expected 'refreshed-key' to be found, but it was missing")}
	if found := store.Contains("keep-key"); !found {t.Errorf("Expected 'keep-key' to be found, but it was missing")}
}

func s(i int)string{return strconv.Itoa(i)}