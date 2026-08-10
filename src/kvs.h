#ifndef KVS_H
#define KVS_H

#include <stddef.h>

typedef struct Node {
    char *key;
    char *value;
    struct Node *next; // Handling collisions
} Node;

typedef struct {
    Node **buckets;
    size_t capacity;
    size_t size;
} KVSStore;

KVSStore* kvs_create(size_t capacity);
void kvs_free(KVSStore *store);

unsigned long hash_key(const char *key, size_t capacity);

#endif