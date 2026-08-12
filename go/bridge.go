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
	"net/http"
	"unsafe"
)


var store *C.KVSStore

func main() {
	// 1. create the store in C
	store = C.kvs_create(1024)
	if store == nil {
		fmt.Println("Failed to create store!")
		return
	}
	defer C.kvs_free(store)

	// 2. POST /set
	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val := r.URL.Query().Get("value")
		if key == "" || val == "" {
			http.Error(w, "Missing key or value", http.StatusBadRequest)
			return
		}

		cKey := C.CString(key)
		cVal := C.CString(val)
		defer C.free(unsafe.Pointer(cKey))
		defer C.free(unsafe.Pointer(cVal))

		C.kvs_set(store, cKey, cVal)
		fmt.Fprintf(w, "Success: Key '%s' stored in C engine!\n", key)
	})

	// 3. GET /get
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing key", http.StatusBadRequest)
			return
		}

		cKey := C.CString(key)
		defer C.free(unsafe.Pointer(cKey))

		cRetVal := C.kvs_get(store, cKey)
		if cRetVal != nil {
			goVal := C.GoString(cRetVal)
			fmt.Fprintf(w, "Value: %s\n", goVal)
		} else {
			http.Error(w, "Key not found", http.StatusNotFound)
		}
	})

	// 4. run server on port 8080
	fmt.Println("🚀 C-Go-KVS Live Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}