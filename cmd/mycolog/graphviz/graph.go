package graphviz

import (
	"fmt"
	"io"
	"os/exec"
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
		ranks[relative.Component.CreatedAt] = append(rank, relative.Component.ID)
	}
	for _, r := range relatives {
		for _, child := range r.Children {
			desc.WriteString(fmt.Sprintf("\t%d -> %d\n", r.Component.ID, child))
		}
	}
	for _, rank := range ranks {
		desc.WriteString("{rank = same;")
		for _, id := range rank {
			desc.WriteString(fmt.Sprintf(" %d;", id))
		}
		desc.WriteString("}\n")
	}
	desc.WriteString("}\n")
	return desc.String()
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
