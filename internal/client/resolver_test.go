// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"reflect"
	"testing"
)

// In a real test, we would mock registry.GetDependencies by injecting an interface
// or replacing the function. Since registry.GetDependencies is a package level function,
// for testing TopologicalSort directly, we can just populate the graph manually.

func TestDependencyGraph_TopologicalSort_Success(t *testing.T) {
	graph := NewDependencyGraph()
	// A depends on B, C
	// B depends on D
	// C depends on D
	// D has no dependencies

	graph.Nodes["A"] = &Node{Name: "A", Dependencies: []string{"B", "C"}}
	graph.Nodes["B"] = &Node{Name: "B", Dependencies: []string{"D"}}
	graph.Nodes["C"] = &Node{Name: "C", Dependencies: []string{"D"}}
	graph.Nodes["D"] = &Node{Name: "D", Dependencies: []string{}}

	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// We expect D to be first. Then B and C (in any order). Then A.
	if len(sorted) != 4 {
		t.Fatalf("expected 4 packages, got %d", len(sorted))
	}
	if sorted[0] != "D" {
		t.Errorf("expected first package to be D, got %s", sorted[0])
	}
	if sorted[3] != "A" {
		t.Errorf("expected last package to be A, got %s", sorted[3])
	}
	
	validOrder1 := []string{"D", "B", "C", "A"}
	validOrder2 := []string{"D", "C", "B", "A"}
	
	if !reflect.DeepEqual(sorted, validOrder1) && !reflect.DeepEqual(sorted, validOrder2) {
		t.Errorf("unexpected order: %v", sorted)
	}
}

func TestDependencyGraph_TopologicalSort_Circular(t *testing.T) {
	graph := NewDependencyGraph()
	// A depends on B
	// B depends on C
	// C depends on A (circular)

	graph.Nodes["A"] = &Node{Name: "A", Dependencies: []string{"B"}}
	graph.Nodes["B"] = &Node{Name: "B", Dependencies: []string{"C"}}
	graph.Nodes["C"] = &Node{Name: "C", Dependencies: []string{"A"}}

	_, err := graph.TopologicalSort()
	if err != ErrCircularDependency {
		t.Fatalf("expected ErrCircularDependency, got %v", err)
	}
}
