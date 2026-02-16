package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onlyboxes/onlyboxes/console/internal/registry"
)

const maxPageSize = 100

type WorkerHandler struct {
	store      *registry.Store
	offlineTTL time.Duration
	nowFn      func() time.Time
}

type workerItem struct {
	NodeID       string                        `json:"node_id"`
	NodeName     string                        `json:"node_name"`
	ExecutorKind string                        `json:"executor_kind"`
	Languages    []registry.LanguageCapability `json:"languages"`
	Labels       map[string]string             `json:"labels"`
	Version      string                        `json:"version"`
	Status       registry.WorkerStatus         `json:"status"`
	RegisteredAt time.Time                     `json:"registered_at"`
	LastSeenAt   time.Time                     `json:"last_seen_at"`
}

type listWorkersResponse struct {
	Items    []workerItem `json:"items"`
	Total    int          `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

func NewWorkerHandler(store *registry.Store, offlineTTL time.Duration) *WorkerHandler {
	return &WorkerHandler{
		store:      store,
		offlineTTL: offlineTTL,
		nowFn:      time.Now,
	}
}

func NewRouter(workerHandler *WorkerHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/api/v1/workers", workerHandler.ListWorkers)
	router.GET("/api/v1/workers/stats", workerHandler.WorkerStats)
	return router
}

func (h *WorkerHandler) ListWorkers(c *gin.Context) {
	page, ok := parsePositiveIntQuery(c, "page", 1)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be a positive integer"})
		return
	}
	pageSize, ok := parsePositiveIntQuery(c, "page_size", 20)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_size must be a positive integer"})
		return
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	status := registry.WorkerStatus(c.DefaultQuery("status", string(registry.StatusAll)))
	if status != registry.StatusAll && status != registry.StatusOnline && status != registry.StatusOffline {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be one of all|online|offline"})
		return
	}

	workers, total := h.store.List(status, page, pageSize, h.nowFn(), h.offlineTTL)
	items := make([]workerItem, 0, len(workers))
	for _, worker := range workers {
		items = append(items, workerItem{
			NodeID:       worker.NodeID,
			NodeName:     worker.NodeName,
			ExecutorKind: worker.ExecutorKind,
			Languages:    worker.Languages,
			Labels:       worker.Labels,
			Version:      worker.Version,
			Status:       worker.Status,
			RegisteredAt: worker.RegisteredAt,
			LastSeenAt:   worker.LastSeenAt,
		})
	}

	c.JSON(http.StatusOK, listWorkersResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func parsePositiveIntQuery(c *gin.Context, key string, defaultValue int) (int, bool) {
	raw := c.Query(key)
	if raw == "" {
		return defaultValue, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}
