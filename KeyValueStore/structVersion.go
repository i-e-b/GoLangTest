package keyvaluestore

import (
	"errors"
	"fmt"
	"sync"
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

// IndependentStore - this struct isn't exported
// it can still be used if returned by a function, but can't be directly new()'d
type IndependentStore struct {
	//OpenClose // Injects a field that is a reference (default nil) to an implementation. Compiler is wonky here, beware.

	// private
	isOpen bool
	coreMap map[StoreKey]StoreValue // interface always acts like a pointer?
	mutex *sync.RWMutex

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

// String satisfies the Stringer interface. It doesn't matter if we use `(receiver *IndependentStore)` or `(receiver IndependentStore)`
func (receiver *IndependentStore) String() string {
	receiver.mutex.RLock()
	defer receiver.mutex.RUnlock()

	return fmt.Sprintf("Key value store (%d keys, is open = %v)", len(receiver.coreMap), receiver.isOpen)
}

// OpenNew is an alternative to `new(IndependentStore)`, used like `keyvaluestore.OpenNew()`
func OpenNew() *IndependentStore {
	iNum++
	store := IndependentStore{
		isOpen: true,
		coreMap: map[StoreKey]StoreValue{},// or `make(map[StoreKey]StoreValue),`, but this is considered 'oldthink'
		InstanceNum: iNum, // you NEED a trailing comma if the closing brace is on a new line
		mutex: &sync.RWMutex{},
	}
	return &store
}

func (receiver *IndependentStore)Open() error {
	// receiver is never null?
	if receiver == nil || receiver.isOpen {return StoreAlreadyOpenError}

	receiver.mutex.Lock()
	defer receiver.mutex.Unlock()

	receiver.isOpen = true
	return nil
}

func (receiver *IndependentStore)Close() error {
	// receiver is never null?
	if receiver == nil || !receiver.isOpen {return StoreNotOpenError}

	receiver.mutex.Lock()
	defer receiver.mutex.Unlock()

	receiver.isOpen = false
	return nil
}

func CloseExisting(store *IndependentStore) error {
	if store == nil || !store.isOpen {return StoreNotOpenError}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.isOpen = false
	return nil
}

func PutValue(store *IndependentStore, key StoreKey, value interface{}) error {
	if store == nil || !store.isOpen {return StoreNotOpenError}
	if store.coreMap == nil {return InvalidStoreError}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.coreMap[key] = &timestampWrapper{
		lastAccess: time.Now(),
		value:      value,
	}

	return nil
}

func (receiver *IndependentStore)Put(key StoreKey, value interface{}) error {
	if receiver == nil || !receiver.isOpen {return StoreNotOpenError}
	if receiver.coreMap == nil {return InvalidStoreError}

	receiver.mutex.Lock()
	defer receiver.mutex.Unlock()

	receiver.coreMap[key] = &timestampWrapper{
		lastAccess: time.Now(),
		value:      value,
	}
	return nil
}

func GetValue(store *IndependentStore, key StoreKey) (interface{}, error){
	if store == nil || !store.isOpen {return "", StoreNotOpenError}
	if store.coreMap == nil {return "", InvalidStoreError}

	store.mutex.RLock() // even though we technically write here; it's a timestamp, so we allow latest-writer-wins
	defer store.mutex.RUnlock()

	value, ok := store.coreMap[key]
	if !ok {return "", KeyNotPresentError}

	value.SetTimestamp(time.Now())
	return fmt.Sprintf("%v",value.GetValue()), nil // seems a bit mental, but is about the only way to cast interface to string
}

func (receiver *IndependentStore)Get(key StoreKey) (interface{}, error){
	if receiver == nil || !receiver.isOpen {return "", StoreNotOpenError}
	if receiver.coreMap == nil {return "", InvalidStoreError}
	value, ok := receiver.coreMap[key]
	if !ok {return "", KeyNotPresentError}

	receiver.mutex.RLock() // even though we technically write here; it's a timestamp, so we allow latest-writer-wins
	defer receiver.mutex.RUnlock()

	value.SetTimestamp(time.Now())

	return value.GetValue(), nil
}

func (receiver *IndependentStore)GetAge(key StoreKey) (time.Time, error){
	if receiver == nil || !receiver.isOpen {return time.Time{}, StoreNotOpenError}
	if receiver.coreMap == nil {return time.Time{}, InvalidStoreError}
	value, ok := receiver.coreMap[key]
	if !ok {return time.Time{}, KeyNotPresentError}

	receiver.mutex.RLock()
	defer receiver.mutex.RUnlock()

	return value.GetTimestamp(), nil
}

func DeleteValue(store *IndependentStore, key StoreKey) error{
	if store == nil || !store.isOpen {return StoreNotOpenError}
	if store.coreMap == nil {return InvalidStoreError}
	if _, ok := store.coreMap[key]; !ok {return KeyNotPresentError}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	delete(store.coreMap, key)
	return nil
}

func (receiver *IndependentStore)Delete(key StoreKey) error{
	if receiver == nil || !receiver.isOpen {return StoreNotOpenError}
	if receiver.coreMap == nil {return InvalidStoreError}
	if _, ok := receiver.coreMap[key]; !ok {return KeyNotPresentError}

	receiver.mutex.Lock()
	defer receiver.mutex.Unlock()

	delete(receiver.coreMap, key)
	return nil
}

func (receiver *IndependentStore)Contains(key StoreKey) bool{
	if receiver == nil || !receiver.isOpen {return false}
	if receiver.coreMap == nil {return false}

	receiver.mutex.RLock()
	defer receiver.mutex.RUnlock()

	_, ok := receiver.coreMap[key]
	return ok
}

func (receiver *IndependentStore) PutWithAge(key StoreKey, value interface{}, timestamp time.Time) error {
	if receiver == nil || !receiver.isOpen {return StoreNotOpenError}
	if receiver.coreMap == nil {return InvalidStoreError}

	receiver.mutex.Lock()
	defer receiver.mutex.Unlock()

	receiver.coreMap[key] = &timestampWrapper{
		lastAccess: timestamp,
		value:      value,
	}
	return nil
}

func (receiver *IndependentStore) EvictOlderThan(timestamp time.Time) {
	if receiver == nil || !receiver.isOpen {return}
	if receiver.coreMap == nil {return}

	receiver.mutex.Lock()
	defer receiver.mutex.Unlock()

	for key, value := range receiver.coreMap {
		realAge := value.GetTimestamp()
		if realAge.After(timestamp) {
			delete(receiver.coreMap, key)
		}
	}
}