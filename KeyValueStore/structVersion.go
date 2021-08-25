package keyvaluestore

import (
	"fmt"
	"time"
)

var iNum=0

type StoreKey string

type StoreValue interface {
	SetTimestamp(t time.Time)
	GetTimestamp() time.Time
	GetValue() interface{}
}

// independentStore - this struct isn't exported
// it can still be used if returned by a function, but can't be directly new()'d
type independentStore struct {
	// private
	isOpen bool
	coreMap map[StoreKey]StoreValue // interface always acts like a pointer?

	// public?
	InstanceNum int
}

type stringWrapper struct{
	lastAccess time.Time
	value string
}
func (receiver *stringWrapper) SetTimestamp(t time.Time) {receiver.lastAccess=t }
func (receiver *stringWrapper) GetTimestamp()time.Time {return receiver.lastAccess}
func (receiver *stringWrapper) GetValue()interface{} {return receiver.value}

func (receiver independentStore) String() string {
	return fmt.Sprintf("Key value store (%d keys, is open = %v)", len(receiver.coreMap), receiver.isOpen)
}

// OpenNew is an alternative to `new(independentStore)`, used like `keyvaluestore.OpenNew()`
func OpenNew() *independentStore {
	iNum++
	store := independentStore{
		isOpen: true,
		coreMap: map[StoreKey]StoreValue{},// or `make(map[StoreKey]StoreValue),`, but this is considered 'oldthink'
		InstanceNum: iNum, // you NEED a trailing comma if the closing brace is on a new line
	}
	return &store
}

func (receiver *independentStore)Open() error {
	// receiver is never null?
	if receiver == nil || receiver.isOpen {return StoreAlreadyOpenError}
	receiver.isOpen = true
	return nil
}

func (receiver *independentStore)Close() error {
	// receiver is never null?
	if receiver == nil || !receiver.isOpen {return StoreNotOpenError}
	receiver.isOpen = false
	return nil
}

func CloseExisting(store *independentStore) error {
	if store == nil || !store.isOpen {return StoreNotOpenError}
	store.isOpen = false
	return nil
}

func PutValue(store *independentStore, key StoreKey, value string) error {
	if store == nil || !store.isOpen {return StoreNotOpenError}
	if store.coreMap == nil {return InvalidStoreError}
	store.coreMap[key] = &stringWrapper{
		lastAccess: time.Now(),
		value:      value,
	}
	return nil
}

func (receiver *independentStore)Put(key StoreKey, value string) error {
	if receiver == nil || !receiver.isOpen {return StoreNotOpenError}
	if receiver.coreMap == nil {return InvalidStoreError}
	receiver.coreMap[key] = &stringWrapper{
		lastAccess: time.Now(),
		value:      value,
	}
	return nil
}

func GetValue(store *independentStore, key StoreKey) (string, error){
	if store == nil || !store.isOpen {return "", StoreNotOpenError}
	if store.coreMap == nil {return "", InvalidStoreError}
	value, ok := store.coreMap[key]
	if !ok {return "", KeyNotPresentError}
	return fmt.Sprintf("%v",value.GetValue()), nil // seems a bit mental, but is about the only way to cast interface to string
}

func (receiver *independentStore)Get(key StoreKey) (string, error){
	if receiver == nil || !receiver.isOpen {return "", StoreNotOpenError}
	if receiver.coreMap == nil {return "", InvalidStoreError}
	value, ok := receiver.coreMap[key]
	if !ok {return "", KeyNotPresentError}
	value.SetTimestamp(time.Now())
	//fmt.Printf("The raw key is %T, %v", value,value) // shows this is a pointer -> "The raw key is *keyvaluestore.stringWrapper, &{{13853528404013489860 2887601 0x62a2e0} CorrectValue}"
	return fmt.Sprintf("%v",value.GetValue()), nil
}

func (receiver *independentStore)GetAge(key StoreKey) (time.Time, error){
	if receiver == nil || !receiver.isOpen {return time.Time{}, StoreNotOpenError}
	if receiver.coreMap == nil {return time.Time{}, InvalidStoreError}
	value, ok := receiver.coreMap[key]
	if !ok {return time.Time{}, KeyNotPresentError}
	value.SetTimestamp(time.Now())
	return value.GetTimestamp(), nil
}

func DeleteValue(store *independentStore, key StoreKey) error{
	if store == nil || !store.isOpen {return StoreNotOpenError}
	if store.coreMap == nil {return InvalidStoreError}
	if _, ok := store.coreMap[key]; !ok {return KeyNotPresentError}
	delete(store.coreMap, key)
	return nil
}

func (receiver *independentStore)Delete(key StoreKey) error{
	if receiver == nil || !receiver.isOpen {return StoreNotOpenError}
	if receiver.coreMap == nil {return InvalidStoreError}
	if _, ok := receiver.coreMap[key]; !ok {return KeyNotPresentError}
	delete(receiver.coreMap, key)
	return nil
}

func (receiver *independentStore)Contains(key StoreKey) bool{
	if receiver == nil || !receiver.isOpen {return false}
	if receiver.coreMap == nil {return false}
	_, ok := receiver.coreMap[key]
	return ok
}

func (receiver *independentStore) PutWithAge(key StoreKey, value string, timestamp time.Time) error {
	if receiver == nil || !receiver.isOpen {return StoreNotOpenError}
	if receiver.coreMap == nil {return InvalidStoreError}
	receiver.coreMap[key] = &stringWrapper{
		lastAccess: timestamp,
		value:      value,
	}
	return nil
}

func (receiver *independentStore) EvictOlderThan(timestamp time.Time) {
	for key, value := range receiver.coreMap {
		realAge := value.GetTimestamp()
		if realAge.After(timestamp) {
			delete(receiver.coreMap, key)
		}
	}
}