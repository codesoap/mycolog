package store

// TODO: Transactions (not _that_ necessary until multiple users use one DB)

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// AddComponent adds a new component. ID and Token will be generated and returned.
// Type and CreatedAt must be specified.
func (db DB) AddComponent(component Component) (id int64, token string, err error) {
	if err = db.validateComponent(component); err != nil {
		return
	}
	token = generateToken()
	query := `INSERT INTO component(type, species, token, createdAt, notes, gone) ` +
		`VALUES (?, ?, ?, ?, ?, ?)`
	var res sql.Result
	res, err = db.Exec(query,
		component.Type,
		component.Species,
		token,
		component.CreatedAt.Format(timeFormat),
		component.Notes,
		component.Gone)
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	return
}

// SetParents stores the relationships between child and its parents
// in the database. An error will be returned if the species of a parent
// does not match the one of the child or if a parent has not been created
// before the child.
func (db DB) SetParents(child int64, parents []int64) error {
	if err := db.checkAdoptability(child, parents); err != nil {
		return err
	}
	deleteQuery := `DELETE FROM relation WHERE child = ?`
	if _, err := db.Exec(deleteQuery, child); err != nil {
		return err
	}
	if len(parents) > 0 {
		insertQuery := `INSERT INTO relation(parent, child) VALUES` +
			` (?, ?)` + strings.Repeat(`, (?, ?)`, len(parents)-1)
		args := make([]interface{}, 0)
		for _, parent := range parents {
			args = append(args, parent, child)
		}
		if _, err := db.Exec(insertQuery, args...); err != nil {
			return err
		}
	}
	return nil
}

// Update updates the component with the given ID. If the ID does not
// exist, an error will be returned. createdAt must be after all its
// parents.
func (db DB) UpdateComponent(id int64, createdAt time.Time, notes string, gone bool) error {
	parentIDs, err := db.GetParents(id)
	if err != nil {
		return err
	}
	parents, err := db.GetComponents(parentIDs)
	if err != nil {
		return err
	}
	for _, parent := range parents {
		if createdAt.Sub(parent.CreatedAt) < 24*time.Hour {
			return fmt.Errorf("the new creation date is too old")
		}
	}

	query := `UPDATE component SET createdAt = ?, notes = ?, gone = ? ` +
		`WHERE id = ? `
	_, err = db.Exec(query,
		createdAt.Format(timeFormat),
		notes,
		gone,
		id)
	return err
}

// UpdateSpecies updates the species of all given ids. Make sure to
// always update all relatives at once to avoid invalid lineages.
func (db DB) UpdateSpecies(ids []int64, species string) error {
	if len(ids) == 0 {
		return nil
	}
	query := `UPDATE component SET species = ? ` +
		`WHERE id IN (?` + strings.Repeat(`, ?`, len(ids)-1) + `)`
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, species)
	for _, id := range ids {
		args = append(args, id)
	}
	_, err := db.Exec(query, args...)
	return err
}

// DeleteComponent deletes a component and all its relationships from
// the database.
func (db DB) DeleteComponent(id int64) error {
	deleteQuery := `DELETE FROM relation WHERE child = ? OR parent = ?`
	if _, err := db.Exec(deleteQuery, id, id); err != nil {
		return err
	}
	_, err := db.Exec(`DELETE FROM component WHERE id = ?`, id)
	return err
}

func (db DB) validateComponent(component Component) error {
	if component.Type != TypeSpores &&
		component.Type != TypeMycelium &&
		component.Type != TypeSpawn &&
		component.Type != TypeGrow {
		return fmt.Errorf("invalid type '%s'", component.Type)
	} else if component.CreatedAt.IsZero() {
		return fmt.Errorf("time of creation not specified")
	}
	return nil
}

func (db DB) checkAdoptability(child int64, parents []int64) error {
	var err error
	var components []Component
	if components, err = db.GetComponents(append(parents, child)); err != nil {
		return err
	}
	parentComponents := components[:len(parents)]
	childComponent := components[len(parents)]
	for _, parentComponent := range parentComponents {
		if parentComponent.Species != childComponent.Species {
			return fmt.Errorf("a parent is not of the same species")
		} else if childComponent.CreatedAt.Sub(parentComponent.CreatedAt) < 24*time.Hour {
			return fmt.Errorf("a parent is too young")
		}
	}
	return nil
}
