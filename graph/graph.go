package graph

import (
	"github.com/codesoap/mycolog/store"
)

// A Relative represents a component together with its children.
type Relative struct {
	Component store.Component
	Parents   []int64
	Children  []int64
}

// GetAllRelatives retrieves all components that are related to the
// component with the given id, including genetically unrelated
// relatives. The component with the given id will be included in the
// results.
func GetAllRelatives(db store.DB, id int64) ([]int64, error) {
	// Implements breadth first search.
	queue := []int64{id}
	visited := make(map[int64]bool)
	for len(queue) > 0 {
		currID := queue[0]
		queue = queue[1:]
		if _, found := visited[currID]; found {
			continue
		}
		visited[currID] = true
		currChildren, err := db.GetChildren(currID)
		if err != nil {
			return nil, err
		}
		queue = append(queue, currChildren...)
		currParents, err := db.GetParents(currID)
		if err != nil {
			return nil, err
		}
		queue = append(queue, currParents...)
	}
	relatives := make([]int64, 0, len(visited))
	for relative := range visited {
		relatives = append(relatives, relative)
	}
	return relatives, nil
}

// GetFullLineage retrieves components which share a genetic connection
// with the component of the given id, including that component.
func GetFullLineage(db store.DB, id int64) ([]Relative, error) {
	roots, err := findRoots(db, id)
	if err != nil {
		return nil, err
	}
	return getAncestors(db, roots)
}

// GetCloseLineage retrieves the parents and all their ancestors
// of the component with the given id, including that component.
func GetCloseLineage(db store.DB, id int64) ([]Relative, error) {
	localRoots, err := db.GetParents(id)
	if err != nil {
		return nil, err
	}
	if len(localRoots) == 0 {
		localRoots = []int64{id}
	}
	return getAncestors(db, localRoots)
}

func findRoots(db store.DB, id int64) ([]int64, error) {
	visited := make(map[int64]bool)
	var queue, roots []int64
	queue = append(queue, id)
	for len(queue) > 0 {
		currID := queue[0]
		queue = queue[1:]
		if _, found := visited[currID]; found {
			continue
		}
		visited[currID] = true
		currParents, err := db.GetParents(currID)
		if err != nil {
			return nil, err
		} else if len(currParents) == 0 {
			roots = append(roots, currID)
		} else {
			queue = append(queue, currParents...)
		}
	}
	return roots, nil
}

func getAncestors(db store.DB, roots []int64) ([]Relative, error) {
	queue := roots[:]
	visited := make(map[int64]bool)
	var relatives []Relative
	for len(queue) > 0 {
		currID := queue[0]
		queue = queue[1:]
		if _, found := visited[currID]; found {
			continue
		}
		visited[currID] = true
		currChildren, err := db.GetChildren(currID)
		if err != nil {
			return nil, err
		}
		currParents, err := db.GetParents(currID)
		if err != nil {
			return nil, err
		}
		component, err := db.GetComponent(currID)
		if err != nil {
			return nil, err
		}
		relative := Relative{Component: component, Parents: currParents, Children: currChildren}
		relatives = append(relatives, relative)
		queue = append(queue, currChildren...)
	}
	return relatives, nil
}
