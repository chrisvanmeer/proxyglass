package server

import (
	"fmt"
	"strings"

	"github.com/chrisvanmeer/proxyglass/internal/config"
	"github.com/chrisvanmeer/proxyglass/internal/prefs"
)

const (
	topologyNodeWidth  = 190.0
	topologyNodeHeight = 70.0
	topologyColumnGap  = 230.0
)

type topologyTarget struct {
	name      string
	frontends []FView
	backends  []BView
}

// buildTopology creates a stable, target-grouped graph from HAProxy stats.
// Frontend-to-backend edges are only added when explicitly configured because
// HAProxy's read-only stats CSV does not expose use_backend routing rules.
func buildTopology(targets []string, frontends []FView, backends []BView, errorsByTarget map[string]error, p *prefs.Prefs, mappings []config.TopologyRoute) TopologyView {
	groups := make([]topologyTarget, 0, len(targets))
	groupIndex := make(map[string]int, len(targets))
	addTarget := func(name string) int {
		if index, ok := groupIndex[name]; ok {
			return index
		}
		groupIndex[name] = len(groups)
		groups = append(groups, topologyTarget{name: name})
		return len(groups) - 1
	}
	for _, target := range targets {
		addTarget(target)
	}
	for _, frontend := range frontends {
		index := addTarget(frontend.Target)
		if frontend.Shown {
			groups[index].frontends = append(groups[index].frontends, frontend)
		}
	}
	for _, backend := range backends {
		index := addTarget(backend.Target)
		if backend.Shown {
			groups[index].backends = append(groups[index].backends, backend)
		}
	}

	graph := TopologyView{Nodes: []TopologyNode{}, Edges: []TopologyEdge{}, Width: 800, Height: 420}
	frontendNodes := make(map[string]TopologyNode)
	backendNodes := make(map[string]TopologyNode)
	left := 40.0
	maxHeight := 420.0
	for _, group := range groups {
		if p != nil && p.TargetFilterActive() && !p.ShowTargets[prefs.KeyTarget(group.name)] {
			continue
		}
		columns := max(len(group.frontends), len(group.backends), 1)
		clusterWidth := max(360.0, float64(columns)*topologyColumnGap)
		targetX := left + (clusterWidth-topologyNodeWidth)/2
		targetStatus, targetState := "ONLINE", "healthy"
		if _, failed := errorsByTarget[group.name]; failed {
			targetStatus, targetState = "UNREACHABLE", "unhealthy"
		}
		targetNode := TopologyNode{
			Kind:       "target",
			Label:      group.name,
			Target:     "HAProxy target",
			Status:     targetStatus,
			StateClass: targetState,
			X:          targetX,
			Y:          32,
		}
		graph.Nodes = append(graph.Nodes, targetNode)

		frontendXs := topologyNodePositions(left, clusterWidth, len(group.frontends))
		for i, frontend := range group.frontends {
			node := TopologyNode{
				Kind:       "frontend",
				Label:      frontend.Name,
				Target:     frontend.Target,
				Status:     frontend.Status,
				StateClass: topologyStatusClass(frontend.Status),
				Anchor:     frontend.Anchor,
				X:          frontendXs[i],
				Y:          158,
			}
			graph.Nodes = append(graph.Nodes, node)
			frontendNodes[topologyKey(frontend.Target, frontend.Name)] = node
			graph.Edges = append(graph.Edges, topologyEdge(targetNode, node, "target"))
		}

		backendXs := topologyNodePositions(left, clusterWidth, len(group.backends))
		for i, backend := range group.backends {
			node := TopologyNode{
				Kind:       "backend",
				Label:      backend.Name,
				Target:     backend.Target,
				Status:     backend.Status,
				StateClass: topologyStatusClass(backend.Status),
				Anchor:     backend.Anchor,
				X:          backendXs[i],
				Y:          300,
			}
			graph.Nodes = append(graph.Nodes, node)
			backendNodes[topologyKey(backend.Target, backend.Name)] = node
			graph.Edges = append(graph.Edges, topologyEdge(targetNode, node, "target"))

			serverY := 442.0
			for _, server := range backend.Servers {
				if !server.Shown {
					continue
				}
				serverNode := TopologyNode{
					Kind:       "server",
					Label:      server.Name,
					Target:     backend.Name,
					Status:     server.Status,
					StateClass: topologyStatusClass(server.Status),
					Anchor:     "server-" + server.Key,
					X:          node.X,
					Y:          serverY,
				}
				graph.Nodes = append(graph.Nodes, serverNode)
				graph.Edges = append(graph.Edges, topologyEdge(node, serverNode, "server"))
				serverY += 94
			}
			maxHeight = max(maxHeight, serverY+20)
		}

		for _, frontend := range group.frontends {
			for _, mapping := range mappings {
				if mapping.Target != group.name || mapping.Frontend != frontend.Name {
					continue
				}
				from, ok := frontendNodes[topologyKey(group.name, frontend.Name)]
				if !ok {
					continue
				}
				for _, backendName := range mapping.Backends {
					to, ok := backendNodes[topologyKey(group.name, backendName)]
					if ok {
						graph.Edges = append(graph.Edges, topologyEdge(from, to, "route"))
					}
				}
			}
		}
		left += clusterWidth + 70
	}
	graph.Width = max(800, int(left-30))
	graph.Height = int(maxHeight)
	return graph
}

func topologyNodePositions(left, width float64, count int) []float64 {
	positions := make([]float64, count)
	if count == 0 {
		return positions
	}
	if count == 1 {
		positions[0] = left + (width-topologyNodeWidth)/2
		return positions
	}
	step := (width - topologyNodeWidth) / float64(count-1)
	for i := range positions {
		positions[i] = left + float64(i)*step
	}
	return positions
}

func topologyEdge(from, to TopologyNode, kind string) TopologyEdge {
	fromX := from.X + topologyNodeWidth/2
	fromY := from.Y + topologyNodeHeight
	toX := to.X + topologyNodeWidth/2
	toY := to.Y
	middleY := (fromY + toY) / 2
	path := fmt.Sprintf("M %.1f %.1f C %.1f %.1f, %.1f %.1f, %.1f %.1f", fromX, fromY, fromX, middleY, toX, middleY, toX, toY)
	return TopologyEdge{Kind: kind, Path: path}
}

func topologyKey(target, name string) string { return target + "\x00" + name }

func topologyStatusClass(status string) string {
	switch strings.ToUpper(status) {
	case "UP", "OPEN", "ONLINE":
		return "healthy"
	case "DOWN", "ERROR", "UNREACHABLE":
		return "unhealthy"
	default:
		return "unknown"
	}
}
