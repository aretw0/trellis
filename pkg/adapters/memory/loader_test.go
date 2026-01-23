package memory_test

import (
	"testing"

	"github.com/aretw0/trellis/pkg/adapters/memory"
	contract "github.com/aretw0/trellis/pkg/ports/tests"
)

func TestInMemoryLoader_Contract(t *testing.T) {
	data := map[string]string{
		"start": "Hello World",
		"end":   "Goodbye",
	}

	// We pass raw bytes map because Loader expects that, but the helper makes strict type check?
	// The helper `GraphLoaderContractTest` expects data as map[string][]byte for comparison.
	// So we need to convert my string map to byte map for the test harness.

	bytesData := make(map[string][]byte)
	for k, v := range data {
		bytesData[k] = []byte(v)
	}

	loader := memory.NewLoader(data)

	contract.GraphLoaderContractTest(t, loader, bytesData)
}
