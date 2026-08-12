#include <stdio.h>
#include "src/kvs.h"

int main() {
    printf("Initializing C-Go-KVS Engine...\n");
    
    KVSStore *store = kvs_create(16);
    if (store) {
        printf("Store created with capacity: %zu\n", store->capacity);
    }
    //set data
    for (int i = 0; i < 5; i++) {
        char key[20];
        char value[20];
        sprintf(key, "username%d", i);
        sprintf(value, "mbrik%d", i);
        if (kvs_set(store, key, value)) {
            printf("Data set successfully!\n");
        } else {
            printf("Failed to set data!\n");
        }
    }
    //get data
    printf("Store final capacity: %zu, size: %zu\n", store->capacity, store->size);
    const char* value = kvs_get(store, "username17");
    if (value) {
        printf("Value: %s\n", value);
    } else {
        printf("Value not found!\n");
    }

    
    kvs_free(store);
    return 0;
}