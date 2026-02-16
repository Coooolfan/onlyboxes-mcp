package registry

import (
	"errors"
	"sort"
	"sync"
	"time"

	registryv1 "github.com/onlyboxes/onlyboxes/api/gen/go/registry/v1"
)

var ErrNodeNotFound = errors.New("worker node not found")

type WorkerStatus string

const (
	StatusAll     WorkerStatus = "all"
	StatusOnline  WorkerStatus = "online"
	StatusOffline WorkerStatus = "offline"
)

type LanguageCapability struct {
	Language string `json:"language"`
	Version  string `json:"version"`
}

type Worker struct {
	NodeID       string
	NodeName     string
	ExecutorKind string
	Languages    []LanguageCapability
	Labels       map[string]string
	Version      string
	RegisteredAt time.Time
	LastSeenAt   time.Time
}

type WorkerView struct {
	Worker
	Status WorkerStatus
}

type WorkerStats struct {
	Total   int
	Online  int
	Offline int
	Stale   int
}

type Store struct {
	mu    sync.RWMutex
	nodes map[string]Worker
}

func NewStore() *Store {
	return &Store{
		nodes: make(map[string]Worker),
	}
}

func (s *Store) Upsert(req *registryv1.RegisterRequest, now time.Time) {
	if req == nil || req.GetNodeId() == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.nodes[req.GetNodeId()]
	registeredAt := now
	if ok {
		registeredAt = existing.RegisteredAt
	}

	s.nodes[req.GetNodeId()] = Worker{
		NodeID:       req.GetNodeId(),
		NodeName:     req.GetNodeName(),
		ExecutorKind: req.GetExecutorKind(),
		Languages:    cloneProtoLanguages(req.GetLanguages()),
		Labels:       cloneMap(req.GetLabels()),
		Version:      req.GetVersion(),
		RegisteredAt: registeredAt,
		LastSeenAt:   now,
	}
}

func (s *Store) Touch(nodeID string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, ok := s.nodes[nodeID]
	if !ok {
		return ErrNodeNotFound
	}
	node.LastSeenAt = now
	s.nodes[nodeID] = node
	return nil
}

func (s *Store) List(status WorkerStatus, page int, pageSize int, now time.Time, offlineTTL time.Duration) ([]WorkerView, int) {
	s.mu.RLock()
	workers := make([]Worker, 0, len(s.nodes))
	for _, node := range s.nodes {
		workers = append(workers, cloneWorker(node))
	}
	s.mu.RUnlock()

	sort.Slice(workers, func(i, j int) bool {
		if workers[i].RegisteredAt.Equal(workers[j].RegisteredAt) {
			return workers[i].NodeID < workers[j].NodeID
		}
		return workers[i].RegisteredAt.Before(workers[j].RegisteredAt)
	})

	filtered := make([]WorkerView, 0, len(workers))
	for _, worker := range workers {
		workerStatus := statusOf(worker.LastSeenAt, now, offlineTTL)
		if status != StatusAll && status != workerStatus {
			continue
		}
		filtered = append(filtered, WorkerView{
			Worker: worker,
			Status: workerStatus,
		})
	}

	total := len(filtered)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	start := (page - 1) * pageSize
	if start >= total {
		return []WorkerView{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return filtered[start:end], total
}

func (s *Store) Stats(now time.Time, offlineTTL time.Duration, staleAfter time.Duration) WorkerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := WorkerStats{}
	for _, worker := range s.nodes {
		stats.Total++
		workerStatus := statusOf(worker.LastSeenAt, now, offlineTTL)
		if workerStatus == StatusOnline {
			stats.Online++
		} else {
			stats.Offline++
		}
		if now.Sub(worker.LastSeenAt) > staleAfter {
			stats.Stale++
		}
	}
	return stats
}

func (s *Store) PruneOffline(now time.Time, offlineTTL time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := 0
	for nodeID, worker := range s.nodes {
		if statusOf(worker.LastSeenAt, now, offlineTTL) == StatusOffline {
			delete(s.nodes, nodeID)
			removed++
		}
	}
	return removed
}

func statusOf(lastSeenAt time.Time, now time.Time, offlineTTL time.Duration) WorkerStatus {
	if now.Sub(lastSeenAt) <= offlineTTL {
		return StatusOnline
	}
	return StatusOffline
}

func cloneWorker(worker Worker) Worker {
	worker.Languages = cloneLanguages(worker.Languages)
	worker.Labels = cloneMap(worker.Labels)
	return worker
}

func cloneProtoLanguages(languages []*registryv1.LanguageCapability) []LanguageCapability {
	if len(languages) == 0 {
		return []LanguageCapability{}
	}
	cloned := make([]LanguageCapability, 0, len(languages))
	for _, language := range languages {
		if language == nil {
			continue
		}
		cloned = append(cloned, LanguageCapability{
			Language: language.GetLanguage(),
			Version:  language.GetVersion(),
		})
	}
	return cloned
}

func cloneLanguages(languages []LanguageCapability) []LanguageCapability {
	if len(languages) == 0 {
		return []LanguageCapability{}
	}
	cloned := make([]LanguageCapability, len(languages))
	copy(cloned, languages)
	return cloned
}

func cloneMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(input))
	for k, v := range input {
		cloned[k] = v
	}
	return cloned
}
