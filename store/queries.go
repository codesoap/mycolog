package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// FIXME: Omit "Get" Prefix in Go fashion.

// A ComponentFilter can contain filter criteria for finding components.
type ComponentFilter struct {
	Types   []ComponentType // If empty, all types will be matched.
	Species []string        // If empty, all species will be matched.
	Since   *time.Time      // If nil, no filtering will occur.
	Until   *time.Time      // If nil, no filtering will occur.
	Gone    *bool           // If nil, no filtering will occur.
}

// GetAllSpecies returns a list of all different species that components
// posses.
func (db DB) GetAllSpecies() ([]string, error) {
	rows, err := db.Query(`SELECT DISTINCT species FROM component`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	allSpecies := make([]string, 0)
	for rows.Next() {
		var species string
		if err := rows.Scan(&species); err != nil {
			return nil, err
		}
		allSpecies = append(allSpecies, species)
	}
	return allSpecies, nil
}

// ComponentsPresent returns true if the database contains at least one
// component.
func (db DB) ComponentsPresent() (bool, error) {
	rows, err := db.Query(`SELECT id FROM component LIMIT 1`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

// GetComponent retrieves the component with the given ID from the
// database. If the component cannot be found, an error will be
// returned.
func (db DB) GetComponent(id int64) (Component, error) {
	ids := make([]int64, 1)
	ids[0] = id
	components, err := db.GetComponents(ids)
	if err != nil {
		return Component{}, err
	}
	return components[0], nil
}

// GetComponents retrieves the components for the given IDs from
// the database. If any component cannot be found, an error will be
// returned.
func (db DB) GetComponents(ids []int64) (components []Component, err error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query := `SELECT id, type, species, token, createdAt, notes, gone ` +
		`FROM component WHERE id IN (?` + strings.Repeat(`, ?`, len(ids)-1) + `)`
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	var rows *sql.Rows
	if rows, err = db.Query(query, args...); err != nil {
		return
	}
	defer rows.Close()
	components, err = getComponentsFromRows(rows)
	if err == nil && len(ids) != len(components) {
		err = fmt.Errorf("component not found")
	}
	return
}

// FindComponents retrieves all components matching the given filter
// from the database.
func (db DB) FindComponents(filter ComponentFilter) ([]Component, error) {
	query := `SELECT id, type, species, token, createdAt, notes, gone ` +
		`FROM component`
	where, args := componentFilterToWhereClause(filter)
	query += where
	var rows *sql.Rows
	query += ` ORDER BY createdAt DESC, id DESC`
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getComponentsFromRows(rows)
}

// GetGrowInfo finds the additional grow information for a component in
// the database. If there is none, the default info will be returned.
func (db DB) GetGrowInfo(id int64) (GrowInfo, error) {
	growInfo := GrowInfo{ID: id}
	query := `SELECT yield, yieldComment FROM grow WHERE id = ?`
	rows, err := db.Query(query, id)
	if err != nil {
		return growInfo, err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&growInfo.Yield, &growInfo.YieldComment)
	}
	return growInfo, err
}

// GetYields finds the yields in milligrams for the components matching
// the given compFilter.
func (db DB) GetYields(compFilter ComponentFilter) (map[int64]*int, error) {
	yields := make(map[int64]*int)
	query := `SELECT grow.id, grow.yield FROM grow ` +
		`INNER JOIN component ON component.id = grow.id`
	where, args := componentFilterToWhereClause(compFilter)
	query += where
	rows, err := db.Query(query, args...)
	if err != nil {
		return yields, err
	}
	defer rows.Close()
	var id int64
	var yield *int
	for rows.Next() {
		err = rows.Scan(&id, &yield)
		if err != nil {
			return yields, err
		}
		yields[id] = yield
	}
	return yields, nil
}

// GetParents finds all parents for the given child in the database.
func (db DB) GetParents(child int64) ([]int64, error) {
	query := `SELECT parent FROM relation WHERE child = ?`
	var rows *sql.Rows
	rows, err := db.Query(query, child)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	parents := make([]int64, 0)
	for rows.Next() {
		var parent int64
		if err := rows.Scan(&parent); err != nil {
			return nil, err
		}
		parents = append(parents, parent)
	}
	return parents, nil
}

// GetChildren finds all children for the given parent in the database.
func (db DB) GetChildren(parent int64) ([]int64, error) {
	query := `SELECT child FROM relation WHERE parent = ?`
	var rows *sql.Rows
	rows, err := db.Query(query, parent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	children := make([]int64, 0)
	for rows.Next() {
		var child int64
		if err := rows.Scan(&child); err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return children, nil
}

func componentFilterToWhereClause(filter ComponentFilter) (string, []any) {
	whereClause := ""
	args := make([]any, 0)
	glue := ` WHERE `
	if len(filter.Types) > 0 {
		whereClause += glue + `component.type IN (?` +
			strings.Repeat(`, ?`, len(filter.Types)-1) +
			`)`
		glue = ` AND `
		for _, arg := range filter.Types {
			args = append(args, arg)
		}
	}
	if len(filter.Species) > 0 {
		whereClause += glue + `component.species IN (?` +
			strings.Repeat(`, ?`, len(filter.Species)-1) +
			`)`
		glue = ` AND `
		for _, arg := range filter.Species {
			args = append(args, arg)
		}
	}
	if filter.Since != nil {
		whereClause += glue + `component.createdAt >= ?`
		glue = ` AND `
		args = append(args, filter.Since.Format(timeFormat))
	}
	if filter.Until != nil {
		whereClause += glue + `component.createdAt <= ?`
		glue = ` AND `
		args = append(args, filter.Until.Format(timeFormat))
	}
	if filter.Gone != nil {
		whereClause += glue + `component.gone = ?`
		args = append(args, *filter.Gone)
	}
	return whereClause, args
}

func getComponentsFromRows(rows *sql.Rows) (components []Component, err error) {
	for rows.Next() {
		var c Component
		var createdAtString string
		err = rows.Scan(&c.ID, &c.Type, &c.Species, &c.Token, &createdAtString,
			&c.Notes, &c.Gone)
		if err != nil {
			return
		}
		c.CreatedAt, err = time.Parse(timeFormat, createdAtString)
		if err != nil {
			return
		}
		components = append(components, c)
	}
	return
}
