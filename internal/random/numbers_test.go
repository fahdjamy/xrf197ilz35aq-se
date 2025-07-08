package random_test

import (
	"bytes"
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"xrf197ilz35aq/internal/random"
)

func TestInt64FromUUID(t *testing.T) {
	t.Run("Generate a set of unique int64 values and check for duplicates", func(t *testing.T) {
		const numValues = 10000
		seenValues := make(map[int64]bool)
		for i := 0; i < numValues; i++ {
			value, err := random.Int64FromUUID()
			assert.NoError(t, err)
			assertValueNotSeen(t, value, &seenValues)
		}
	})

	t.Run("Concurrently generate values in multiple goroutines and check for duplicates", func(t *testing.T) {
		const numOfGoRoutines = 10
		const valuesPerGoroutine = 1000

		// Duplicate Detection
		seenValues := make(map[int64]bool)

		mutex := sync.Mutex{}
		var wg sync.WaitGroup
		wg.Add(numOfGoRoutines)

		for i := 0; i < numOfGoRoutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < valuesPerGoroutine; j++ {
					value, err := random.Int64FromUUID()
					assert.NoError(t, err)

					mutex.Lock()
					if seenValues[value] {
						t.Errorf("Duplicate value generated in goroutine: %d", value)
					}
					seenValues[value] = true
					mutex.Unlock()
				}
			}()
		}
		wg.Wait()
	})

	t.Run("Fails with an invalid reader", func(t *testing.T) {
		// Simulate a failure in UUID generation (replace rand.Reader with a faulty reader)
		originalReader := rand.Reader

		// This will cause uuid.NewRandomFromReader to fail
		// Unsafe, remember to always restore to originalReader
		rand.Reader = bytes.NewReader(nil)
		defer func() { rand.Reader = originalReader }()

		_, err := random.Int64FromUUID()
		assert.Error(t, err)
	})
}

func TestInt64(t *testing.T) {
	t.Run("it generates unique positive values", func(t *testing.T) {
		const numValues = 10000
		seenValues := make(map[int64]bool)
		for i := 0; i < numValues; i++ {
			positiveInt64 := random.PositiveInt64()
			if positiveInt64 < 0 {
				t.Errorf("Negative value generated: %d", positiveInt64)
			}
			assertValueNotSeen(t, positiveInt64, &seenValues)
		}
	})

	t.Run("it creates unique values in when called concurrently", func(t *testing.T) {
		const numOfGoRoutines = 10
		const valuesPerGoroutine = 1000
		seenValues := make(map[int64]bool)
		mutex := sync.Mutex{}
		var wg sync.WaitGroup
		wg.Add(numOfGoRoutines)
		for i := 0; i < numOfGoRoutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < valuesPerGoroutine; j++ {
					value := random.PositiveInt64()
					mutex.Lock()
					assertValueNotSeen(t, value, &seenValues)
					mutex.Unlock()
				}
			}()
		}
		wg.Wait()
	})
}

func assertValueNotSeen(t *testing.T, value int64, seenValues *map[int64]bool) {
	t.Helper()
	// Dereference the pointer to get the actual map
	theMap := *seenValues
	if theMap[value] {
		t.Errorf("Duplicate value generated: %d", value)
	}
	theMap[value] = true
}
