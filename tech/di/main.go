package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phamnam2003/challenges/tech/di/db"
	"github.com/phamnam2003/challenges/tech/di/handler"
	"github.com/phamnam2003/challenges/tech/di/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewRouter(h *handler.HealthHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	h.Register(r)

	return r
}

func NewHTTPServer(
	lc fx.Lifecycle,
	router *gin.Engine,
	log *zap.Logger,
) *http.Server {

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("http server starting", zap.String("addr", srv.Addr))

			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatal("http server crashed", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("http server stopping")

			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			return srv.Shutdown(shutdownCtx)
		},
	})

	return srv
}

func New() *fx.App {
	return fx.New(
		fx.WithLogger(logger.NewFxEventLogger),
		fx.Provide(
			logger.NewZap,
			db.NewStorage,
			db.NewCache,
			handler.NewHealthHandler,
			NewHTTPServer,
			NewRouter,
		),
		fx.Invoke(func(*http.Server) {}),
	)
}

func main() {
	New().Run()
}
