// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/snowdreamtech/unistack/internal/registry"
)

var (
	ErrCircularDependency = errors.New("circular dependency detected in package graph")
)

// Node represents a package in the dependency graph
type Node struct {
	Name         string
	Dependencies []string // packages this node depends on
}

// DependencyGraph represents the directed acyclic graph of package dependencies
type DependencyGraph struct {
	Nodes map[string]*Node
}

// NewDependencyGraph creates an empty dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes: make(map[string]*Node),
	}
}

// BuildGraph recursively builds the dependency graph for a target package
func (g *DependencyGraph) BuildGraph(ctx context.Context, db *sql.DB, targetPkg string) error {
	// If the package is already in the graph, skip to prevent infinite loops during graph building
	if _, exists := g.Nodes[targetPkg]; exists {
		return nil
	}

	// Fetch dependencies for this package
	deps, err := registry.GetDependencies(ctx, db, targetPkg)
	if err != nil {
		return fmt.Errorf("failed to fetch dependencies for %s: %w", targetPkg, err)
	}

	// Add node to graph
	g.Nodes[targetPkg] = &Node{
		Name:         targetPkg,
		Dependencies: deps,
	}

	// Recursively build graph for dependencies
	for _, dep := range deps {
		if err := g.BuildGraph(ctx, db, dep); err != nil {
			return err
		}
	}

	return nil
}

// TopologicalSort performs a topological sort using Kahn's algorithm.
// It returns a slice of package names in the order they should be installed.
func (g *DependencyGraph) TopologicalSort() ([]string, error) {
	// Calculate in-degree (number of packages that depend on this node)
	// For installation, if A depends on B, we want B to be installed before A.
	// So we create a graph where edges point from dependency to dependent.
	// B -> A means A depends on B.
	// A node with in-degree 0 means nothing depends on it? No, wait.
	// We want to install nodes that have NO dependencies first.
	// So out-degree in the original graph (edges from dependent to dependency).
	// Let's compute the number of dependencies for each node.

	inDegree := make(map[string]int)
	graph := make(map[string][]string) // dependency -> list of dependents

	// Initialize maps
	for name := range g.Nodes {
		inDegree[name] = 0
		graph[name] = []string{}
	}

	// Populate graph and in-degrees
	for name, node := range g.Nodes {
		// name depends on node.Dependencies
		// For our sorting, we want to pop nodes with 0 dependencies first.
		inDegree[name] = len(node.Dependencies)
		for _, dep := range node.Dependencies {
			// dep has a dependent: name
			graph[dep] = append(graph[dep], name)
		}
	}

	// Queue for nodes with 0 dependencies
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	var sorted []string

	// Process queue
	for len(queue) > 0 {
		// Pop
		curr := queue[0]
		queue = queue[1:]

		sorted = append(sorted, curr)

		// Decrease in-degree for dependents
		for _, dependent := range graph[curr] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// If sorted length != number of nodes, we have a cycle
	if len(sorted) != len(g.Nodes) {
		return nil, ErrCircularDependency
	}

	return sorted, nil
}
