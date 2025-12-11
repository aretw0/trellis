package compiler

import (
	"encoding/json"
	"fmt"

	"github.com/aretw0/trellis/pkg/domain"
)

// Parser is responsible for converting raw bytes into a Node.
type Parser struct{}

// NewParser creates a new parser instance.
func NewParser() *Parser {
	return &Parser{}
}

// Parse takes the raw content and tries to decode it into a Node.
// For MVP, we assume the content is JSON.
func (p *Parser) Parse(data []byte) (*domain.Node, error) {
	var node domain.Node
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("failed to parse node: %w", err)
	}
	// Basic validation
	if node.ID == "" {
		return nil, fmt.Errorf("node missing ID")
	}
	return &node, nil
}
