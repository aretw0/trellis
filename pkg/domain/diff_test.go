package domain

import (
	"reflect"
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
				Context:    map[string]any{},
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
				Context:       map[string]any{},
				HistoryParams: &HistoryDelta{Appended: []string{"next"}},
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

func equalPtr[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
