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
	mu    sync.RWMutex // لحماية محرك C من تضارب الـ Goroutines أثناء العمليات المتزامنة
)

type SetPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func main() {
	// 1. إنشاء مخزن البيانات (KVS Store) في محرك C
	store = C.kvs_create(1024)
	if store == nil {
		fmt.Println("Failed to create KVS store in C engine!")
		return
	}
	defer C.kvs_free(store)

	mux := http.NewServeMux()

	// 2. مسار إضافة أو تعديل البيانات: POST /set
	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
			return
		}

		var payload SetPayload
		// قراءة البيانات من body بصيغة JSON أولاً، وإن تعذر فمن الـ Query Parameters
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

		// قفل الكتابة لمنع التضارب في الذاكرة (Race Conditions)
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

	// 3. مسار جلب البيانات: GET /get
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

		// قفل القراءة لتسمح بالقراءات المتزامنة بأمان
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

	// 4. تشغيل خادم HTTP على المنفذ 8080
	fmt.Println("🚀 C-Go-KVS Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
