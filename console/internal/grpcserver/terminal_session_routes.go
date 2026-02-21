package grpcserver

import (
	"strings"
	"time"
)

type terminalSessionRoute struct {
	NodeID         string
	LastUsedUnixMs int64
}

func (s *RegistryService) bindTerminalSessionRoute(sessionID string, nodeID string, now time.Time) {
	if s == nil {
		return
	}
	normalizedSessionID := strings.TrimSpace(sessionID)
	normalizedNodeID := strings.TrimSpace(nodeID)
	if normalizedSessionID == "" || normalizedNodeID == "" {
		return
	}

	nowUnixMs := routeNowUnixMs(now)
	s.terminalRoutesMu.Lock()
	defer s.terminalRoutesMu.Unlock()

	existing, exists := s.terminalSessionToNode[normalizedSessionID]
	if exists && existing.NodeID != normalizedNodeID {
		previousIndex := s.terminalNodeToSessionIDIndex[existing.NodeID]
		if previousIndex != nil {
			delete(previousIndex, normalizedSessionID)
			if len(previousIndex) == 0 {
				delete(s.terminalNodeToSessionIDIndex, existing.NodeID)
			}
		}
	}

	s.terminalSessionToNode[normalizedSessionID] = terminalSessionRoute{
		NodeID:         normalizedNodeID,
		LastUsedUnixMs: nowUnixMs,
	}
	index := s.terminalNodeToSessionIDIndex[normalizedNodeID]
	if index == nil {
		index = make(map[string]struct{})
		s.terminalNodeToSessionIDIndex[normalizedNodeID] = index
	}
	index[normalizedSessionID] = struct{}{}
}

func (s *RegistryService) reserveTerminalSessionRoute(sessionID string, preferredNodeID string, now time.Time) (string, bool) {
	if s == nil {
		return "", false
	}
	normalizedSessionID := strings.TrimSpace(sessionID)
	normalizedNodeID := strings.TrimSpace(preferredNodeID)
	if normalizedSessionID == "" || normalizedNodeID == "" {
		return "", false
	}

	nowUnixMs := routeNowUnixMs(now)
	s.terminalRoutesMu.Lock()
	defer s.terminalRoutesMu.Unlock()

	existing, exists := s.terminalSessionToNode[normalizedSessionID]
	if exists {
		existing.LastUsedUnixMs = nowUnixMs
		s.terminalSessionToNode[normalizedSessionID] = existing
		return existing.NodeID, false
	}

	s.terminalSessionToNode[normalizedSessionID] = terminalSessionRoute{
		NodeID:         normalizedNodeID,
		LastUsedUnixMs: nowUnixMs,
	}
	index := s.terminalNodeToSessionIDIndex[normalizedNodeID]
	if index == nil {
		index = make(map[string]struct{})
		s.terminalNodeToSessionIDIndex[normalizedNodeID] = index
	}
	index[normalizedSessionID] = struct{}{}
	return normalizedNodeID, true
}

// touchTerminalSessionRoute returns the mapped node and refreshes LastUsedUnixMs.
func (s *RegistryService) touchTerminalSessionRoute(sessionID string, now time.Time) (string, bool) {
	if s == nil {
		return "", false
	}
	normalizedSessionID := strings.TrimSpace(sessionID)
	if normalizedSessionID == "" {
		return "", false
	}

	nowUnixMs := routeNowUnixMs(now)
	s.terminalRoutesMu.Lock()
	defer s.terminalRoutesMu.Unlock()

	route, ok := s.terminalSessionToNode[normalizedSessionID]
	if !ok || strings.TrimSpace(route.NodeID) == "" {
		return "", false
	}
	route.LastUsedUnixMs = nowUnixMs
	s.terminalSessionToNode[normalizedSessionID] = route
	return route.NodeID, true
}

func (s *RegistryService) clearTerminalSessionRoute(sessionID string, expectedNodeID string) {
	if s == nil {
		return
	}
	normalizedSessionID := strings.TrimSpace(sessionID)
	if normalizedSessionID == "" {
		return
	}
	normalizedExpectedNodeID := strings.TrimSpace(expectedNodeID)

	s.terminalRoutesMu.Lock()
	defer s.terminalRoutesMu.Unlock()

	route, ok := s.terminalSessionToNode[normalizedSessionID]
	if !ok {
		return
	}
	if normalizedExpectedNodeID != "" && route.NodeID != normalizedExpectedNodeID {
		return
	}

	delete(s.terminalSessionToNode, normalizedSessionID)
	index := s.terminalNodeToSessionIDIndex[route.NodeID]
	if index == nil {
		return
	}
	delete(index, normalizedSessionID)
	if len(index) == 0 {
		delete(s.terminalNodeToSessionIDIndex, route.NodeID)
	}
}

func (s *RegistryService) clearTerminalSessionRoutesByNode(nodeID string) {
	if s == nil {
		return
	}
	normalizedNodeID := strings.TrimSpace(nodeID)
	if normalizedNodeID == "" {
		return
	}

	s.terminalRoutesMu.Lock()
	defer s.terminalRoutesMu.Unlock()

	index := s.terminalNodeToSessionIDIndex[normalizedNodeID]
	if index == nil {
		return
	}
	for sessionID := range index {
		route, ok := s.terminalSessionToNode[sessionID]
		if !ok || route.NodeID != normalizedNodeID {
			continue
		}
		delete(s.terminalSessionToNode, sessionID)
	}
	delete(s.terminalNodeToSessionIDIndex, normalizedNodeID)
}

func (s *RegistryService) pruneExpiredTerminalSessionRoutes(now time.Time) int {
	if s == nil {
		return 0
	}
	ttl := s.terminalRouteTTL
	if ttl <= 0 {
		return 0
	}
	nowUnixMs := routeNowUnixMs(now)
	expireBefore := nowUnixMs - ttl.Milliseconds()

	removed := 0
	s.terminalRoutesMu.Lock()
	defer s.terminalRoutesMu.Unlock()

	for sessionID, route := range s.terminalSessionToNode {
		if route.LastUsedUnixMs > expireBefore {
			continue
		}
		delete(s.terminalSessionToNode, sessionID)
		index := s.terminalNodeToSessionIDIndex[route.NodeID]
		if index != nil {
			delete(index, sessionID)
			if len(index) == 0 {
				delete(s.terminalNodeToSessionIDIndex, route.NodeID)
			}
		}
		removed++
	}
	return removed
}

func (s *RegistryService) maybePruneTerminalSessionRoutes(now time.Time) {
	if s == nil {
		return
	}
	nowUnixMs := routeNowUnixMs(now)
	minIntervalMs := terminalRoutePruneMinInterval.Milliseconds()

	for {
		last := s.lastTerminalRoutePruneUnixMs.Load()
		if last > 0 && nowUnixMs-last < minIntervalMs {
			return
		}
		if s.lastTerminalRoutePruneUnixMs.CompareAndSwap(last, nowUnixMs) {
			break
		}
	}
	s.pruneExpiredTerminalSessionRoutes(now)
}

func routeNowUnixMs(now time.Time) int64 {
	if now.IsZero() {
		return time.Now().UnixMilli()
	}
	return now.UnixMilli()
}
