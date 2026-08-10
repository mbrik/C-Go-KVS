#include <stdio.h>
#include "src/kvs.h"

int main() {
    printf("Initializing C-Go-KVS Engine...\n");
    
    KVSStore *store = kvs_create(16);
    if (store) {
        printf("Store created with capacity: %zu\n", store->capacity);
    }

    kvs_free(store);
    return 0;
}