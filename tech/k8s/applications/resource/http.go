package resource

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type HttpServer struct {
	router  *gin.Engine
	options HttpOpts
}

func (s *HttpServer) setupRoutes() {
	var router *gin.Engine
	router = gin.Default()

	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowCredentials: false,
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Accept-Language"},
	}))

	router.GET("/health-check", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"service": s.options.ServiceName,
		})
	})

	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"service": s.options.ServiceName,
			"error":   "not found route",
		})
	})

	s.router = router
}

func (s *HttpServer) Run(ctx context.Context, wg *errgroup.Group) {
	httpServer := &http.Server{
		Addr:           s.options.Addr,
		Handler:        s.router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 2 << 20, // 2 MiB
	}

	wg.Go(func() error {
		log.Printf("[%s] starting HTTP server on %s", s.options.ServiceName, s.options.Addr)
		err := httpServer.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}

			log.Printf("[%s] cannot start HTTP server on %s: %v", s.options.ServiceName, s.options.Addr, err)
			return err
		}

		return nil
	})

	wg.Go(func() error {
		<-ctx.Done()
		log.Printf("[%s] graceful shutdown http server on %s", s.options.ServiceName, s.options.Addr)

		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.Printf("[%s] failed to shutdown HTTP server on %s: %v", s.options.ServiceName, s.options.Addr, err)
			return err
		}

		log.Printf("[%s] HTTP server on %s was stopped", s.options.ServiceName, s.options.Addr)
		return nil
	})
}

func NewHttpServer(opts ...Options) *HttpServer {
	s := &HttpServer{}
	for _, opt := range opts {
		opt(&s.options)
	}

	s.setupRoutes()

	return s
}
