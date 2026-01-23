package memory_test

import (
	"testing"

	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/ports"
)

func TestMemoryStore_Contract(t *testing.T) {
	store := memory.NewStore()
	ports.RunStateStoreContract(t, store)
}
