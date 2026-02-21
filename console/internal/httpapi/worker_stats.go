package httpapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onlyboxes/onlyboxes/console/internal/grpcserver"
)

const defaultStaleAfterSec = 30

type InflightStatsProvider interface {
	InflightStats() []grpcserver.WorkerInflightSnapshot
}

type workerStatsResponse struct {
	Total         int       `json:"total"`
	Online        int       `json:"online"`
	Offline       int       `json:"offline"`
	Stale         int       `json:"stale"`
	StaleAfterSec int       `json:"stale_after_sec"`
	GeneratedAt   time.Time `json:"generated_at"`
}

type capabilityInflightJSON struct {
	Name        string `json:"name"`
	Inflight    int    `json:"inflight"`
	MaxInflight int    `json:"max_inflight"`
}

type workerInflightJSON struct {
	NodeID       string                   `json:"node_id"`
	Capabilities []capabilityInflightJSON `json:"capabilities"`
}

type workerInflightResponse struct {
	Workers     []workerInflightJSON `json:"workers"`
	GeneratedAt time.Time            `json:"generated_at"`
}

func (h *WorkerHandler) WorkerStats(c *gin.Context) {
	staleAfterSec, ok := parsePositiveIntQuery(c, "stale_after_sec", defaultStaleAfterSec)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stale_after_sec must be a positive integer"})
		return
	}

	now := h.nowFn()
	stats := h.store.Stats(now, h.offlineTTL, time.Duration(staleAfterSec)*time.Second)
	c.JSON(http.StatusOK, workerStatsResponse{
		Total:         stats.Total,
		Online:        stats.Online,
		Offline:       stats.Offline,
		Stale:         stats.Stale,
		StaleAfterSec: staleAfterSec,
		GeneratedAt:   now,
	})
}

func (h *WorkerHandler) WorkerInflight(c *gin.Context) {
	if h.inflightStats == nil {
		c.JSON(http.StatusOK, workerInflightResponse{
			Workers:     []workerInflightJSON{},
			GeneratedAt: h.nowFn(),
		})
		return
	}

	snapshots := h.inflightStats.InflightStats()
	workers := make([]workerInflightJSON, len(snapshots))
	for i, snap := range snapshots {
		entries := make([]capabilityInflightJSON, len(snap.Capabilities))
		for j, entry := range snap.Capabilities {
			entries[j] = capabilityInflightJSON{
				Name:        entry.Name,
				Inflight:    entry.Inflight,
				MaxInflight: entry.MaxInflight,
			}
		}
		workers[i] = workerInflightJSON{
			NodeID:       snap.NodeID,
			Capabilities: entries,
		}
	}
	c.JSON(http.StatusOK, workerInflightResponse{
		Workers:     workers,
		GeneratedAt: h.nowFn(),
	})
}
