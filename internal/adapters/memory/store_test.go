package memory_test

import (
	"testing"

	"github.com/aretw0/trellis/internal/adapters/memory"
	"github.com/aretw0/trellis/pkg/ports"
)

func TestMemoryStore_Contract(t *testing.T) {
	store := memory.New()
	ports.RunStateStoreContract(t, store)
}
