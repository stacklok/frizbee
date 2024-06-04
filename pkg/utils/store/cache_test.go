package store

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type Cacher interface {
	Store(key, value string)
	Load(key string) (string, bool)
}

// TestCacher tests the creation and basic functionality of both refCacher and unsafeCacher.
func TestCacher(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		cacher      Cacher
		key         string
		storeValue  string
		loadKey     string
		expectedVal string
		expectFound bool
	}{
		{
			name:        "RefCacher store and load existing key",
			cacher:      NewRefCacher(),
			key:         "key1",
			storeValue:  "value1",
			loadKey:     "key1",
			expectedVal: "value1",
			expectFound: true,
		},
		{
			name:        "RefCacher load non-existing key",
			cacher:      NewRefCacher(),
			key:         "key1",
			storeValue:  "value1",
			loadKey:     "key2",
			expectedVal: "",
			expectFound: false,
		},
		{
			name:        "UnsafeCacher store and load existing key",
			cacher:      NewUnsafeCacher(),
			key:         "key1",
			storeValue:  "value1",
			loadKey:     "key1",
			expectedVal: "value1",
			expectFound: true,
		},
		{
			name:        "UnsafeCacher load non-existing key",
			cacher:      NewUnsafeCacher(),
			key:         "key1",
			storeValue:  "value1",
			loadKey:     "key2",
			expectedVal: "",
			expectFound: false,
		},
	}

	for _, tt := range testCases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.cacher.Store(tt.key, tt.storeValue)
			val, ok := tt.cacher.Load(tt.loadKey)
			require.Equal(t, tt.expectFound, ok)
			require.Equal(t, tt.expectedVal, val)
		})
	}
}

// TestConcurrency tests the thread-safety of refCacher.
func TestConcurrency(t *testing.T) {
	t.Parallel()

	cacher := NewRefCacher()
	iterations := 1000
	done := make(chan bool)

	// Concurrently store values
	for i := 0; i < iterations; i++ {
		go func(i int) {
			key := fmt.Sprintf("key%d", i)
			value := fmt.Sprintf("value%d", i)
			cacher.Store(key, value)
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish storing
	for i := 0; i < iterations; i++ {
		<-done
	}

	// Concurrently load values
	for i := 0; i < iterations; i++ {
		go func(i int) {
			key := fmt.Sprintf("key%d", i)
			val, ok := cacher.Load(key)
			expectedVal := fmt.Sprintf("value%d", i)
			require.True(t, ok)
			require.Equal(t, expectedVal, val)
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish loading
	for i := 0; i < iterations; i++ {
		<-done
	}
}
