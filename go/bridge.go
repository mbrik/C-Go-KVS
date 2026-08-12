package main

/*
#cgo CFLAGS: -I../src
#cgo LDFLAGS: -L../build -lkvs
#include "kvs.h"
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"unsafe"
)

var (
	store *C.KVSStore
	mu    sync.RWMutex // Protects C store from concurrent goroutine data races
)

type SetPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func main() {
	// 1. Create the KVS store in the C engine
	store = C.kvs_create(1024)
	if store == nil {
		fmt.Println("Failed to create KVS store in C engine!")
		return
	}
	defer C.kvs_free(store)

	mux := http.NewServeMux()

	// 2. POST /set - Insert or update key-value pair
	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
			return
		}

		var payload SetPayload
		// Read JSON body first; fallback to URL query parameters for backward compatibility
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Key == "" {
			payload.Key = r.URL.Query().Get("key")
			payload.Value = r.URL.Query().Get("value")
		}

		if payload.Key == "" || payload.Value == "" {
			http.Error(w, "Missing key or value", http.StatusBadRequest)
			return
		}

		cKey := C.CString(payload.Key)
		cVal := C.CString(payload.Value)
		defer C.free(unsafe.Pointer(cKey))
		defer C.free(unsafe.Pointer(cVal))

		// Acquire write lock to prevent race conditions in C engine
		mu.Lock()
		res := C.kvs_set(store, cKey, cVal)
		mu.Unlock()

		if res == 0 {
			http.Error(w, "Failed to store item in C engine", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"success","key":"%s"}`, payload.Key)
	})

	// 3. GET /get - Retrieve value by key
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed. Use GET", http.StatusMethodNotAllowed)
			return
		}

		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
			return
		}

		cKey := C.CString(key)
		defer C.free(unsafe.Pointer(cKey))

		// Acquire read lock for safe concurrent access
		mu.RLock()
		cRetVal := C.kvs_get(store, cKey)
		mu.RUnlock()

		if cRetVal != nil {
			goVal := C.GoString(cRetVal)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"key":"%s","value":"%s"}`, key, goVal)
		} else {
			http.Error(w, "Key not found", http.StatusNotFound)
		}
	})

	// 4. Start HTTP server on port 8080
	fmt.Println("🚀 C-Go-KVS Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
