package random

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	mutex     sync.Mutex
	lastValue int64
)

func PositiveInt64() int64 {
	// Used to ensure thread-safety.
	// If multiple goroutines call this function simultaneously,
	// the mutex prevents them from interfering with each other and potentially generating the same value
	mutex.Lock()
	defer mutex.Unlock()

	// Get current timestamp in nanoseconds
	now := time.Now().UnixNano()

	// Ensure uniqueness even if called multiple times in the same nanosecond
	if now <= lastValue {
		now = lastValue + 1
	}
	lastValue = now

	// Generate a random int64 to add further uniqueness
	// The maximum value for rand.Int is adjusted to 1<<62 - 1 to ensure the random part is always positive.
	randomPart, _ := rand.Int(rand.Reader, big.NewInt(1<<31-1)) // Max positive int64

	// Clear the most significant 32 bits of the timestamp to ensure it's positive after the shift
	now &= 1<<32 - 1

	// Combine timestamp and random part
	// The timestamp is shifted left by 31 bits instead of 32.
	// This leaves the most significant bit free, guaranteeing the final uniqueValue is always positive.
	uniqueValue := (now << 31) | randomPart.Int64()

	return uniqueValue
}

func Int64FromUUID() (int64, error) {
	mutex.Lock()
	defer mutex.Unlock()

	// Generates a Version 4 UUID (random) using the cryptographically secure random number generator
	uuidStr, err := uuid.NewRandomFromReader(rand.Reader)
	if err != nil {
		return 0, fmt.Errorf("failed to generate UUID: %v", err)
	}

	// Convert UUID to byte slice
	uuidBytes := uuidStr[:]

	// Extract the first 8 bytes of the UUID and convert to int64
	var uniqueInt64 int64

	// Extracts the first 8 bytes (64 bits) from the UUID and
	// interprets them as an int64 value using BigEndian byte order
	// Assuming BigEndian for most systems

	// UUIDs are 128-bit (16-bytes), we extract 8 bytes (64 bits) to ensure
	// that the data fits perfectly into an int64 variable without any truncation or loss of information.
	err = binary.Read(bytes.NewReader(uuidBytes), binary.BigEndian, &uniqueInt64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert UUID to int64: %v", err)
	}

	return uniqueInt64, nil
}
