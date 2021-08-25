package keyvaluestore

import (
	"errors"
	"fmt"
	"time"
)

var iNum=0

type StoreKey string

type Valuable interface {
	GetValue() interface{}
}

type StoreValue interface {
	Valuable // embedding interface in another
	SetTimestamp(t time.Time)
	GetTimestamp() time.Time
}

type OpenClose interface {
	Open() error
	Close() error
}

func Describe(thing interface{}) string {
	// demonstrate type switching and casting from interface to known type
	// you can also do
	//     value,ok := thing.(MyType)
	// which will give ok==false if the cast is bad.
	// If you do
	//     value := thing.(MyType)
	// and the cast is bad, you get a panic.
	switch v := thing.(type) {
	case StoreValue:
		sv := thing.(StoreValue)
		return fmt.Sprintf("Store value '%v', last accessed %v", sv.GetValue(), sv.GetTimestamp())
	case StoreKey:
		return fmt.Sprintf("Store Key ['%v']", thing.(StoreKey))
	case string:
		return fmt.Sprintf("\"%s\"", thing.(string))
	default:
		return fmt.Sprintf("I don't know about type %T!\n", v)
	}
}

var StoreAlreadyOpenError = errors.New("the store is already open")
var StoreNotOpenError = errors.New("the store is not open")
var InvalidStoreError = errors.New("the store is invalid")
var KeyNotPresentError = errors.New("the given key is not present in this store")

// independentStore - this struct isn't exported
// it can still be used if returned by a function, but can't be directly new()'d
type independentStore struct {
	OpenClose // Injects a field that is a reference (default nil) to an implementation. Compiler is wonky here, beware.

	// private
	isOpen bool
	coreMap map[StoreKey]StoreValue // interface always acts like a pointer?

	// public?
	InstanceNum int
}

type timestampWrapper struct{
	lastAccess time.Time
	value interface{}
}
func (receiver *timestampWrapper) SetTimestamp(t time.Time) {receiver.lastAccess=t }
func (receiver *timestampWrapper) GetTimestamp()time.Time   {return receiver.lastAccess}
func (receiver *timestampWrapper) GetValue()interface{}     {return receiver.value}

// String satisfies the Stringer interface. It doesn't matter if we use `(receiver *independentStore)` or `(receiver independentStore)`
func (receiver *independentStore) String() string {
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

func PutValue(store *independentStore, key StoreKey, value interface{}) error {
	if store == nil || !store.isOpen {return StoreNotOpenError}
	if store.coreMap == nil {return InvalidStoreError}
	store.coreMap[key] = &timestampWrapper{
		lastAccess: time.Now(),
		value:      value,
	}
	return nil
}

func (receiver *independentStore)Put(key StoreKey, value interface{}) error {
	if receiver == nil || !receiver.isOpen {return StoreNotOpenError}
	if receiver.coreMap == nil {return InvalidStoreError}
	receiver.coreMap[key] = &timestampWrapper{
		lastAccess: time.Now(),
		value:      value,
	}
	return nil
}

func GetValue(store *independentStore, key StoreKey) (interface{}, error){
	if store == nil || !store.isOpen {return "", StoreNotOpenError}
	if store.coreMap == nil {return "", InvalidStoreError}
	value, ok := store.coreMap[key]
	if !ok {return "", KeyNotPresentError}
	return fmt.Sprintf("%v",value.GetValue()), nil // seems a bit mental, but is about the only way to cast interface to string
}

func (receiver *independentStore)Get(key StoreKey) (interface{}, error){
	if receiver == nil || !receiver.isOpen {return "", StoreNotOpenError}
	if receiver.coreMap == nil {return "", InvalidStoreError}
	value, ok := receiver.coreMap[key]
	if !ok {return "", KeyNotPresentError}
	value.SetTimestamp(time.Now())
	//fmt.Printf("The raw key is %T, %v", value,value) // shows this is a pointer -> "The raw key is *keyvaluestore.timestampWrapper, &{{13853528404013489860 2887601 0x62a2e0} CorrectValue}"
	return value.GetValue(), nil
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

func (receiver *independentStore) PutWithAge(key StoreKey, value interface{}, timestamp time.Time) error {
	if receiver == nil || !receiver.isOpen {return StoreNotOpenError}
	if receiver.coreMap == nil {return InvalidStoreError}
	receiver.coreMap[key] = &timestampWrapper{
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