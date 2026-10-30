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
		targetNode := newTopologyNode("target", group.name, "HAProxy target", targetStatus, targetState, "tcp", "", targetX, 32)
		graph.Nodes = append(graph.Nodes, targetNode)

		frontendXs := topologyFrontendPositions(left, clusterWidth, group, backendNodes, mappings)
		for i, frontend := range group.frontends {
			node := newTopologyNode("frontend", frontend.Name, frontend.Target, frontend.Status, topologyStatusClass(frontend.Status), "http", frontend.Anchor, frontendXs[i], 158)
			graph.Nodes = append(graph.Nodes, node)
			frontendNodes[topologyKey(frontend.Target, frontend.Name)] = node
			graph.Edges = append(graph.Edges, topologyEdge(targetNode, node, "target"))
		}

		backendXs := topologyNodePositions(left, clusterWidth, len(group.backends))
		for i, backend := range group.backends {
			protocol := backendProtocol(backend)
			node := newTopologyNode("backend", backend.Name, backend.Target, backend.Status, topologyStatusClass(backend.Status), protocol, backend.Anchor, backendXs[i], 300)
			graph.Nodes = append(graph.Nodes, node)
			backendNodes[topologyKey(backend.Target, backend.Name)] = node
			graph.Edges = append(graph.Edges, topologyEdge(targetNode, node, "target"))

			serverY := 442.0
			for _, server := range backend.Servers {
				if !server.Shown {
					continue
				}
				serverNode := newTopologyNode("server", server.Name, backend.Name, server.Status, topologyStatusClass(server.Status), protocol, "server-"+server.Key, node.X, serverY)
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

const topologyLabelMaxRunes = 20

func newTopologyNode(kind, label, target, status, stateClass, protocol, anchor string, x, y float64) TopologyNode {
	displayLabel, labelTruncated := truncateTopologyText(label)
	displayMeta, metaTruncated := truncateTopologyText(target + " · " + status)
	return TopologyNode{
		Kind:           kind,
		Label:          label,
		DisplayLabel:   displayLabel,
		LabelTruncated: labelTruncated,
		Target:         target,
		Status:         status,
		DisplayMeta:    displayMeta,
		MetaTruncated:  metaTruncated,
		StateClass:     stateClass,
		Protocol:       protocol,
		Anchor:         anchor,
		X:              x,
		Y:              y,
	}
}

func truncateTopologyText(value string) (string, bool) {
	runes := []rune(value)
	if len(runes) <= topologyLabelMaxRunes {
		return value, false
	}
	return string(runes[:topologyLabelMaxRunes]) + "...", true
}

func topologyFrontendPositions(left, width float64, group topologyTarget, backendNodes map[string]TopologyNode, mappings []config.TopologyRoute) []float64 {
	positions := make([]float64, len(group.frontends))
	if len(group.frontends) == 0 {
		return positions
	}
	// Determine connected backends per frontend.
	connected := make([][]float64, len(group.frontends))
	for i, frontend := range group.frontends {
		for _, mapping := range mappings {
			if mapping.Target != group.name || mapping.Frontend != frontend.Name {
				continue
			}
			for _, backendName := range mapping.Backends {
				if bn, ok := backendNodes[topologyKey(group.name, backendName)]; ok {
					connected[i] = append(connected[i], bn.X)
				}
			}
		}
	}
	// Frontends with connections are centered over their connected backends.
	// Frontends without connections are spread across the cluster.
	unconnected := []int{}
	for i := range group.frontends {
		if len(connected[i]) == 0 {
			unconnected = append(unconnected, i)
		}
	}
	// Spread unconnected frontends in remaining space.
	spreadPositions := topologyNodePositions(left, width, len(unconnected))
	ui := 0
	for i := range group.frontends {
		if len(connected[i]) > 0 {
			minX, maxX := connected[i][0], connected[i][0]
			for _, x := range connected[i] {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
			}
			center := (minX + maxX + topologyNodeWidth) / 2
			positions[i] = center - topologyNodeWidth/2
			// clamp
			if positions[i] < left {
				positions[i] = left
			}
			if positions[i] > left+width-topologyNodeWidth {
				positions[i] = left + width - topologyNodeWidth
			}
		} else {
			positions[i] = spreadPositions[ui]
			ui++
		}
	}
	// Resolve overlaps: if two frontends end up too close, nudge.
	for i := 1; i < len(positions); i++ {
		if positions[i]-positions[i-1] < topologyNodeWidth+20 {
			positions[i] = positions[i-1] + topologyNodeWidth + 20
		}
	}
	return positions
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

func backendProtocol(backend BView) string {
	for _, server := range backend.Servers {
		check := strings.ToUpper(server.CheckStatus)
		if strings.Contains(check, "L7") {
			return "http"
		}
		if strings.Contains(check, "L4") {
			return "tcp"
		}
	}
	return "unknown"
}

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
