package keyvaluestore

import "errors"

var KeyNotPresentError = errors.New("the key is not present in the store")
var StoreNotOpenError = errors.New("the store is not open")
var StoreAlreadyOpenError = errors.New("the store is already open")
var InvalidStoreError = errors.New("the store is not valid")


var isOpen = false
var coreMap = map[string]string{}

// HardReset is for testing only
func HardReset(){
	isOpen = false
}

func Open() error{
	if isOpen {return StoreAlreadyOpenError}
	isOpen = true
	return nil
}

func Close() error{
	if !isOpen {return StoreNotOpenError}
	isOpen=false
	return nil
}

func Put(key string, value string) error{
	if !isOpen {return StoreNotOpenError}
	coreMap[key] = value
	return nil
}

func Get(key string) (string, error){
	if !isOpen {return "", StoreNotOpenError}
	value, ok := coreMap[key]
	if !ok {return "", KeyNotPresentError}
	return value, nil
}

func Delete(key string) error{
	if !isOpen {return StoreNotOpenError}
	if _, ok := coreMap[key]; !ok {return KeyNotPresentError}
	delete(coreMap, key)
	return nil
}