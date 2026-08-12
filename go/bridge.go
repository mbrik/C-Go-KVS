package main

/*
#cgo CFLAGS: -I../src
#cgo LDFLAGS: -L../build -lkvs
#include "kvs.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func main() {
	fmt.Println("Initializing C Core Engine from Go...")

	
	capacity := C.size_t(16)
	store := C.kvs_create(capacity)
	if store == nil {
		fmt.Println("Failed to create store!")
		return
	}
	defer C.kvs_free(store)

	fmt.Println("Store created successfully in C from Go!")

	
	cKey := C.CString("framework")
	cVal := C.CString("Go-C-Interop")
	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cVal))

	C.kvs_set(store, cKey, cVal)
	fmt.Println("Data inserted via Go -> C successfully!")

	
	cRetVal := C.kvs_get(store, cKey)
	if cRetVal != nil {
		
		goVal := C.GoString(cRetVal)
		fmt.Printf("Retrieved from C engine -> Key: framework, Value: %s\n", goVal)
	} else {
		fmt.Println("Key not found!")
	}
}