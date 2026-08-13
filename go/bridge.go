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
	"os"
	"sync"
	"unsafe"
)

var (
	store *C.KVSStore
	mu    sync.RWMutex // Protects C store from concurrent goroutine data races
)

const defaultDataFile = "data.kvs"

type SetPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type FilePayload struct {
	Filename string `json:"filename"`
}

func main() {
	// 1. Create the KVS store in the C engine
	store = C.kvs_create(1024)
	if store == nil {
		fmt.Println("Failed to create KVS store in C engine!")
		return
	}
	defer C.kvs_free(store)

	// Auto-load existing data from file if present
	if _, err := os.Stat(defaultDataFile); err == nil {
		cFile := C.CString(defaultDataFile)
		if C.kvs_load(store, cFile) == 1 {
			fmt.Println("📦 Automatically loaded existing data from data.kvs")
		}
		C.free(unsafe.Pointer(cFile))
	}

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
		if res == 1 {
			// Auto-save to data.kvs for persistence across restarts
			cFile := C.CString(defaultDataFile)
			C.kvs_save(store, cFile)
			C.free(unsafe.Pointer(cFile))
		}
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

	// 4. POST /save - Save store to file persistence
	mux.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
			return
		}

		var payload FilePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Filename == "" {
			payload.Filename = r.URL.Query().Get("filename")
		}
		if payload.Filename == "" {
			payload.Filename = "data.kvs"
		}

		cFilename := C.CString(payload.Filename)
		defer C.free(unsafe.Pointer(cFilename))

		mu.RLock()
		res := C.kvs_save(store, cFilename)
		mu.RUnlock()

		if res == 0 {
			http.Error(w, "Failed to save store to file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"success","filename":"%s"}`, payload.Filename)
	})

	// 5. POST /load - Load store from file persistence
	mux.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
			return
		}

		var payload FilePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Filename == "" {
			payload.Filename = r.URL.Query().Get("filename")
		}
		if payload.Filename == "" {
			payload.Filename = "data.kvs"
		}

		cFilename := C.CString(payload.Filename)
		defer C.free(unsafe.Pointer(cFilename))

		mu.Lock()
		res := C.kvs_load(store, cFilename)
		mu.Unlock()

		if res == 0 {
			http.Error(w, "Failed to load store from file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"success","filename":"%s"}`, payload.Filename)
	})

	// 6. Start HTTP server on port 8080
	fmt.Println("🚀 C-Go-KVS Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
