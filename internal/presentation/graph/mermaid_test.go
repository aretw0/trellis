package graph_test

import (
	"strings"
	"testing"

	"github.com/aretw0/trellis/internal/presentation/graph"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestGenerateMermaid(t *testing.T) {
	tests := []struct {
		name     string
		nodes    []domain.Node
		contains []string
	}{
		{
			name: "Start Node Shape",
			nodes: []domain.Node{
				{ID: "start", Type: domain.NodeTypeText},     // ID="start" trigger
				{ID: "explicit", Type: domain.NodeTypeStart}, // Type="start" trigger
			},
			contains: []string{
				"start((\"start\"))",
				"explicit((\"explicit\"))",
			},
		},
		{
			name: "Tool Node Shape",
			nodes: []domain.Node{
				{ID: "my_tool", Type: domain.NodeTypeTool},
			},
			contains: []string{
				"my_tool[[\"my_tool\"]]",
			},
		},
		{
			name: "Input Node Shape",
			nodes: []domain.Node{
				{ID: "q1", Type: domain.NodeTypeQuestion},
			},
			contains: []string{
				"q1[/\"q1\"/]",
			},
		},
		{
			name: "ID Sanitization",
			nodes: []domain.Node{
				{ID: "path/to/file.md"},
				{ID: "hyphen-ated"},
			},
			contains: []string{
				"path_to_file_md[\"path/to/file.md\"]",
				"hyphen_ated[\"hyphen-ated\"]",
			},
		},
		{
			name: "Transition Escaping",
			nodes: []domain.Node{
				{
					ID: "A",
					Transitions: []domain.Transition{
						{ToNodeID: "B", Condition: `input == "yes"`},
					},
				},
			},
			contains: []string{
				`-- "input == 'yes'" -->`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := graph.GenerateMermaid(tt.nodes)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("GenerateMermaid() = \n%v\nWant substring: %v", got, want)
				}
			}
		})
	}
}
