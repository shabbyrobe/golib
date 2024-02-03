package graph

import (
	"fmt"
	"strings"
)

func Dot[T ID](g *Digraph[T]) (string, error) {
	var sb strings.Builder
	sb.WriteString("digraph D {\n")
	for _, vertex := range g.vertices {
		sb.WriteString(fmt.Sprintf("  %q\n", fmt.Sprintf("%v", vertex.ID)))
	}
	for _, edge := range g.edges {
		sb.WriteString(fmt.Sprintf("  %q -> %q\n", fmt.Sprintf("%v", edge.From), fmt.Sprintf("%v", edge.To)))
	}
	sb.WriteString("}\n")
	return sb.String(), nil
}
