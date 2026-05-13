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
	router      *gin.Engine
	serviceName string
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
			"service": s.serviceName,
		})
	})

	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"service": s.serviceName,
			"error":   "not found route",
		})
	})

	s.router = router
}

func (s *HttpServer) Run(ctx context.Context, wg *errgroup.Group, addr string) {
	httpServer := &http.Server{
		Addr:           addr,
		Handler:        s.router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 2 << 20, // 2 MiB
	}

	wg.Go(func() error {
		log.Printf("[%s] starting HTTP server on %s", s.serviceName, addr)
		err := httpServer.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}

			log.Printf("[%s] cannot start HTTP server on %s: %v", s.serviceName, addr, err)
			return err
		}

		return nil
	})

	wg.Go(func() error {
		<-ctx.Done()
		log.Printf("[%s] graceful shutdown http server on %s", s.serviceName, addr)

		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.Printf("[%s] failed to shutdown HTTP server on %s: %v", s.serviceName, addr, err)
			return err
		}

		log.Printf("[%s] HTTP server on %s was stopped", s.serviceName, addr)
		return nil
	})
}

func NewHttpServer(serviceName string) *HttpServer {
	s := &HttpServer{
		serviceName: serviceName,
	}

	s.setupRoutes()

	return s
}
