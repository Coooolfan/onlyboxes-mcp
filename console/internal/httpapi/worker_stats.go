package httpapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const defaultStaleAfterSec = 30

type workerStatsResponse struct {
	Total         int       `json:"total"`
	Online        int       `json:"online"`
	Offline       int       `json:"offline"`
	Stale         int       `json:"stale"`
	StaleAfterSec int       `json:"stale_after_sec"`
	GeneratedAt   time.Time `json:"generated_at"`
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
