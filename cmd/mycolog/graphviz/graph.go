package graphviz

import (
	"fmt"
	"io"
	"os/exec"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/codesoap/mycolog/graph"
	"github.com/codesoap/mycolog/store"
)

// Render takes a set of relatives and renders them into HTML code, that
// will display a clickable image.
func Render(relatives []graph.Relative, selectedID int64) (string, error) {
	// TODO: Own error when 'dot' not found?
	graphDescription := getGraphDescription(relatives, selectedID)
	return toHTMLImage(graphDescription)
}

func getGraphDescription(relatives []graph.Relative, selectedID int64) string {
	ranks := make(map[time.Time][]int64)
	var desc strings.Builder
	desc.WriteString("digraph family_tree {\n")
	for _, relative := range relatives {
		desc.WriteString(getNodeDesc(relative, selectedID))
		rank := ranks[relative.Component.CreatedAt]
		newrank := false
		// If a child and parent have the same date, we don't want to put them
		// on the same rank
		for _, parent := range relative.Parents {
			if slices.Contains(rank, parent) {
				newrank = true
				break
			}
		}
		if newrank == true {
			// since the date only contains the Y-M-D and the rest of the time
			// structure is zero'd, we can force a new rank without impacting
			// other dates by adding some a duration that's less than a whole
			// day. We'll just use 1 ns for simplicity.
			midrank := ranks[relative.Component.CreatedAt.Add(1)]
			ranks[relative.Component.CreatedAt.Add(1)] = append(midrank, relative.Component.ID)
		} else {
			ranks[relative.Component.CreatedAt] = append(rank, relative.Component.ID)
		}
	}
	for _, r := range relatives {
		for _, child := range r.Children {
			desc.WriteString(fmt.Sprintf("\t%d -> %d\n", r.Component.ID, child))
		}
	}
	for i, rank := range toSortedRankKeys(ranks) {
		// Create "ordering node" for this rank:
		desc.WriteString(fmt.Sprintf("\to%d [style=invis width=0.01 fontsize=1]\n", i))
		if i > 0 {
			desc.WriteString(fmt.Sprintf("\to%d -> o%d [style=invis]\n", i-1, i))
		}

		desc.WriteString("{rank = same;")
		desc.WriteString(fmt.Sprintf(" o%d;", i))
		for _, id := range ranks[rank] {
			desc.WriteString(fmt.Sprintf(" %d;", id))
		}
		desc.WriteString("}\n")
	}
	desc.WriteString("}\n")
	return desc.String()
}

func toSortedRankKeys(in map[time.Time][]int64) []time.Time {
	keys := make([]time.Time, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Before(keys[j]) })
	return keys
}

func getNodeDesc(relative graph.Relative, selectedID int64) string {
	comp := relative.Component
	var color string
	if comp.ID == selectedID {
		color = `fillcolor="#a273ff" fontcolor="white"`
	} else {
		switch relative.Component.Type {
		case store.TypeSpores:
			color = `fillcolor="black" fontcolor="white"`
		case store.TypeMycelium:
			color = `fillcolor="white"`
		case store.TypeSpawn:
			color = `fillcolor="lightgray"`
		case store.TypeGrow:
			color = `fillcolor="gray" fontcolor="white"`
		default:
			panic("unknown component type")
		}
	}
	createdAt := comp.CreatedAt.Format("2006-01-02")
	label := ""
	if relative.Component.Gone {
		label = "† "
	}
	switch relative.Component.Type {
	case store.TypeSpores:
		label += fmt.Sprintf(`Spores %s\n%s`, comp.Token, createdAt)
	case store.TypeMycelium:
		label += fmt.Sprintf(`Myc. %s\n%s`, comp.Token, createdAt)
	case store.TypeSpawn:
		label += fmt.Sprintf(`Spawn %s\n%s`, comp.Token, createdAt)
	case store.TypeGrow:
		label += fmt.Sprintf(`Grow %s\n%s`, comp.Token, createdAt)
	default:
		panic("unknown component type")
	}
	format := "\t%d [style=\"filled\" %s URL=\"/component/%d\" label=\"%s\"]\n"
	return fmt.Sprintf(format, comp.ID, color, comp.ID, label)
}

func toHTMLImage(graphDescription string) (string, error) {
	cmd := exec.Command("dot", "-Tsvg", "-Tcmapx")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, graphDescription)
	}()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	out, err := io.ReadAll(stdout)
	if err != nil {
		return "", err
	}
	return `<img usemap="#family_tree">` + string(out) + `</img>`, cmd.Wait()
}
