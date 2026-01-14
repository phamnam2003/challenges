package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewHealthHandler(db *pgxpool.Pool, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

func (h *HealthHandler) Register(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		ctx := context.Background()

		if err := h.db.Ping(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"db": "down"})
			return
		}

		if err := h.rdb.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"redis": "down"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
