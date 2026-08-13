#include "kvs.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

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

    for (size_t i = 0; i < store->capacity; i++) {
        Node *current = store->buckets[i];
        while (current != NULL) {
            Node *next = current->next;
            free(current->key);
            free(current->value);
            free(current);
            current = next;
        }
    }

    free(store->buckets);
    free(store);
}

// Hashing function
unsigned long hash_key(const char *key, size_t capacity) {
    if (!key || capacity == 0) return 0;
    unsigned long hash = 5381;
    int c;

    while ((c = *key++)) {
        // hash * 33 + c
        hash = ((hash << 5) + hash) + c; 
    }

    return hash % capacity; 
}

// Helper function to resize and rehash the table when load factor exceeds threshold
static int kvs_resize(KVSStore *store) {
    size_t new_capacity = store->capacity * 2;
    Node **new_buckets = (Node**)calloc(new_capacity, sizeof(Node*));
    if (!new_buckets) return 0;

    for (size_t i = 0; i < store->capacity; i++) {
        Node *current = store->buckets[i];
        while (current != NULL) {
            Node *next = current->next;
            unsigned long new_index = hash_key(current->key, new_capacity);
            current->next = new_buckets[new_index];
            new_buckets[new_index] = current;
            current = next;
        }
    }

    free(store->buckets);
    store->buckets = new_buckets;
    store->capacity = new_capacity;
    return 1;
}

// Set function
int kvs_set(KVSStore *store, const char *key, const char *value) {
    if (!store || !key || !value) return 0;

    unsigned long index = hash_key(key, store->capacity);

    // Check if key already exists, update value if so
    Node *current = store->buckets[index];
    while (current != NULL) {
        if (strcmp(current->key, key) == 0) {
            char *new_val = strdup(value);
            if (!new_val) return 0;
            free(current->value);
            current->value = new_val;
            return 1;
        }
        current = current->next;
    }

    // Check capacity and double size if limit reached
    if (store->size >= store->capacity) {
        if (!kvs_resize(store)) {
            return 0; // Allocation failed during resize
        }
        // Recalculate index after resizing
        index = hash_key(key, store->capacity);
    }

    Node *new_node = (Node*)malloc(sizeof(Node));
    if (!new_node) return 0; 

    new_node->key = strdup(key);
    new_node->value = strdup(value);
    if (!new_node->key || !new_node->value) {
        free(new_node->key);
        free(new_node->value);
        free(new_node);
        return 0;
    }

    new_node->next = store->buckets[index];
    store->buckets[index] = new_node;
    store->size++;
    return 1; 
}

const char* kvs_get(KVSStore *store, const char *key) {
    if (!store || !key) return NULL;

    unsigned long index = hash_key(key, store->capacity);
    Node *current = store->buckets[index];

    while (current != NULL) {
        if (strcmp(current->key, key) == 0) {
            return current->value; 
        }
        current = current->next;
    }

    return NULL; 
}

int kvs_save(KVSStore *store, const char *filename) {
    if (!store || !filename) return 0;

    FILE *file = fopen(filename, "w");
    if (!file) return 0;

    for (size_t i = 0; i < store->capacity; i++) {
        Node *current = store->buckets[i];
        while (current != NULL) {
            fprintf(file, "%s=%s\n", current->key, current->value);
            current = current->next;
        }
    }

    fclose(file);
    return 1;
}

int kvs_load(KVSStore *store, const char *filename) {
    if (!store || !filename) return 0;

    FILE *file = fopen(filename, "r");
    if (!file) return 0;

    char line[1024];
    while (fgets(line, sizeof(line), file)) {
        line[strcspn(line, "\r\n")] = 0;
        char *eq = strchr(line, '=');
        if (eq) {
            *eq = '\0';
            char *key = line;
            char *value = eq + 1;
            kvs_set(store, key, value);
        }
    }

    fclose(file);
    return 1;
}