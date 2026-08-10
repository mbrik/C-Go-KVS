#include "kvs.h"
#include <stdio.h>
#include <stdlib.h>

KVSStore* kvs_create(size_t capacity) {
    KVSStore *store = (KVSStore*)malloc(sizeof(KVSStore));
    if (!store) return NULL;

    store->capacity = capacity;
    store->size = 0;
    store->buckets = (Node**)calloc(capacity, sizeof(Node*));

    if (!store->buckets) {
        free(store);
        return NULL;
    }

    return store;
}

void kvs_free(KVSStore *store) {
    if (!store) return;
    free(store->buckets);
    free(store);
    printf("Memory freed successfully!\n");
}
 
//hashing function
unsigned long hash_key(const char *key, size_t capacity) {
    unsigned long hash = 5381;
    int c;

    while ((c = *key++)) {
        // hash * 33 + c
        hash = ((hash << 5) + hash) + c; 
    }

    return hash % capacity; 
}