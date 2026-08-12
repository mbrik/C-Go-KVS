# Compiler and compilation flags
CC = gcc
CFLAGS = -Wall -g -Isrc

# Directory paths
SRC_DIR = src
BUILD_DIR = build

# Source and Object files
C_SOURCES = $(SRC_DIR)/kvs.c
C_OBJECTS = $(BUILD_DIR)/kvs.o

# Default target when running 'make'
all: directories libkvs main

# Create the build directory if it doesn't exist
directories:
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)


# Compile kvs.c into an object file
$(BUILD_DIR)/kvs.o: $(SRC_DIR)/kvs.c $(SRC_DIR)/kvs.h
	$(CC) $(CFLAGS) -c $(SRC_DIR)/kvs.c -o $@

# Create a static library (libkvs.a) for Go to consume
libkvs: $(C_OBJECTS)
	ar rcs $(BUILD_DIR)/libkvs.a $(C_OBJECTS)

# Compile the main test application main.c
main: main.c libkvs
	$(CC) $(CFLAGS) main.c $(BUILD_DIR)/libkvs.a -o $(BUILD_DIR)/main.exe

# Clean up build artifacts
clean:
	@if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
