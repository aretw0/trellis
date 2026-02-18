package domain

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDiff(t *testing.T) {
	active := StatusActive
	term := StatusTerminated

	tests := []struct {
		name     string
		old      *State
		new      *State
		wantDiff *StateDiff // nil means we expect no diff or Empty diff effectively
	}{
		{
			name: "Initial Load (Old is Nil)",
			old:  nil,
			new: &State{
				SessionID:     "sess-1",
				CurrentNodeID: "start",
				Status:        StatusActive,
				Context:       map[string]any{"a": 1},
				History:       []string{"start"},
			},
			wantDiff: &StateDiff{
				SessionID:     "sess-1",
				CurrentNodeID: &[]string{"start"}[0],
				Status:        &active,
				Context:       map[string]any{"a": 1},
				HistoryParams: &HistoryDelta{Appended: []string{"start"}},
			},
		},
		{
			name: "No Changes",
			old: &State{
				SessionID:     "sess-1",
				CurrentNodeID: "start",
				Status:        StatusActive,
				Context:       map[string]any{"a": 1},
				History:       []string{"start"},
			},
			new: &State{
				SessionID:     "sess-1",
				CurrentNodeID: "start",
				Status:        StatusActive,
				Context:       map[string]any{"a": 1},
				History:       []string{"start"},
			},
			wantDiff: nil,
		},
		{
			name: "Status Change & Terminated",
			old: &State{
				SessionID:     "sess-1",
				CurrentNodeID: "end",
				Status:        StatusActive,
			},
			new: &State{
				SessionID:     "sess-1",
				CurrentNodeID: "end",
				Status:        StatusTerminated,
				Terminated:    true,
			},
			wantDiff: &StateDiff{
				SessionID:  "sess-1",
				Status:     &term,
				Terminated: &[]bool{true}[0],
				Context:    nil,
			},
		},
		{
			name: "Context Added & Modified",
			old: &State{
				SessionID:     "sess-1",
				CurrentNodeID: "mid",
				Context:       map[string]any{"a": 1, "b": "old"},
			},
			new: &State{
				SessionID:     "sess-1",
				CurrentNodeID: "mid",
				Context:       map[string]any{"a": 1, "b": "new", "c": true},
			},
			wantDiff: &StateDiff{
				SessionID: "sess-1",
				Context:   map[string]any{"b": "new", "c": true},
			},
		},
		{
			name: "History Append",
			old: &State{
				SessionID:     "sess-1",
				CurrentNodeID: "start",
				History:       []string{"start"},
			},
			new: &State{
				SessionID:     "sess-1",
				CurrentNodeID: "next",
				History:       []string{"start", "next"},
			},
			wantDiff: &StateDiff{
				SessionID:     "sess-1",
				CurrentNodeID: &[]string{"next"}[0],
				Context:       nil,
				HistoryParams: &HistoryDelta{Appended: []string{"next"}},
			},
		},
		{
			name: "Context Deletion",
			old: &State{
				Context: map[string]any{"a": 1, "b": 2},
			},
			new: &State{
				Context: map[string]any{"a": 1},
			},
			wantDiff: &StateDiff{
				Context: map[string]any{"b": nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Diff(tt.old, tt.new)
			if tt.wantDiff == nil {
				if got != nil {
					t.Errorf("Diff() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatalf("Diff() = nil, want %v", tt.wantDiff)
			}

			if got.SessionID != tt.wantDiff.SessionID {
				t.Errorf("Diff().SessionID = %v, want %v", got.SessionID, tt.wantDiff.SessionID)
			}
			// Check Context
			if !reflect.DeepEqual(got.Context, tt.wantDiff.Context) {
				t.Errorf("Diff().Context = %v, want %v", got.Context, tt.wantDiff.Context)
			}
			// Check History
			if !reflect.DeepEqual(got.HistoryParams, tt.wantDiff.HistoryParams) {
				t.Errorf("Diff().HistoryParams = %v, want %v", got.HistoryParams, tt.wantDiff.HistoryParams)
			}
			// Check CurrentNodeID
			if !equalPtr(got.CurrentNodeID, tt.wantDiff.CurrentNodeID) {
				t.Errorf("Diff().CurrentNodeID = %v, want %v", got.CurrentNodeID, tt.wantDiff.CurrentNodeID)
			}
		})
	}
}

func TestDiffJSONSerialization(t *testing.T) {
	t.Run("Empty Context Omitted", func(t *testing.T) {
		s1 := &State{Context: map[string]any{"a": 1}}
		s2 := &State{Context: map[string]any{"a": 1}} // No change
		diff := Diff(s1, s2)

		if diff != nil {
			bytes, _ := json.Marshal(diff)
			if strings.Contains(string(bytes), `"context"`) {
				t.Errorf("JSON should not contain 'context' when empty, got: %s", string(bytes))
			}
		}
	})

	t.Run("Deletions as Null", func(t *testing.T) {
		s1 := &State{Context: map[string]any{"a": 1, "b": 2}}
		s2 := &State{Context: map[string]any{"a": 1}} // 'b' deleted
		diff := Diff(s1, s2)

		if diff == nil {
			t.Fatal("Expected diff, got nil")
		}

		bytes, _ := json.Marshal(diff)
		if !strings.Contains(string(bytes), `"b":null`) {
			t.Errorf("JSON should contain 'b':null for deletion, got: %s", string(bytes))
		}
	})
}

func equalPtr[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
