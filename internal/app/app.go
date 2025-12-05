package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"m3m/internal/config"
	"m3m/internal/domain"
	"m3m/internal/handler"
	"m3m/internal/middleware"
	"m3m/internal/plugin"
	"m3m/internal/repository"
	"m3m/internal/runtime"
	"m3m/internal/service"
	"m3m/web"
)

func NewLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	switch cfg.Logging.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}

func NewGinEngine(cfg *config.Config) *gin.Engine {
	if cfg.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	return r
}

func RegisterRoutes(
	r *gin.Engine,
	cfg *config.Config,
	logger *slog.Logger,
	authMiddleware *middleware.AuthMiddleware,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	projectHandler *handler.ProjectHandler,
	goalHandler *handler.GoalHandler,
	pipelineHandler *handler.PipelineHandler,
	storageHandler *handler.StorageHandler,
	modelHandler *handler.ModelHandler,
	envHandler *handler.EnvironmentHandler,
	runtimeHandler *handler.RuntimeHandler,
	widgetHandler *handler.WidgetHandler,
) {
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")

	// Public routes
	authHandler.Register(api)

	// Protected routes
	authHandler.RegisterProtected(api, authMiddleware)
	userHandler.Register(api, authMiddleware)
	projectHandler.Register(api, authMiddleware)
	goalHandler.Register(api, authMiddleware)
	pipelineHandler.Register(api, authMiddleware)
	storageHandler.Register(api, authMiddleware)
	modelHandler.Register(api, authMiddleware)
	envHandler.Register(api, authMiddleware)
	runtimeHandler.Register(api, authMiddleware)
	widgetHandler.Register(api, authMiddleware)

	// Public project routes (at root level, not under /api)
	runtimeHandler.RegisterPublicRoutes(r)

	// Register UI static routes
	registerUIRoutes(r, cfg, logger)
}

func registerUIRoutes(r *gin.Engine, cfg *config.Config, logger *slog.Logger) {
	// Check if UI is available
	if !web.HasUI() {
		logger.Warn("UI not available, skipping static routes registration")
		return
	}

	// Get index.html with injected config
	indexHTML, err := web.GetIndexHTML(cfg)
	if err != nil {
		logger.Error("Failed to get index.html", "error", err)
		return
	}

	// Get static filesystem
	staticFS, err := web.GetFileSystem()
	if err != nil {
		logger.Error("Failed to get static filesystem", "error", err)
		return
	}

	// Serve static assets
	r.StaticFS("/assets", staticFS)

	// SPA fallback - all non-API routes return index.html
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip API routes
		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// Skip health check
		if path == "/health" {
			return
		}

		// Return index.html for SPA routing
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	logger.Info("UI routes registered")
}

func StartServer(lc fx.Lifecycle, r *gin.Engine, cfg *config.Config, logger *slog.Logger, runtimeManager *runtime.Manager) {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: r,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting server", "addr", server.Addr)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("Server error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping server")
			logger.Info("Stopping all runtimes...")
			runtimeManager.StopAll()
			return server.Shutdown(ctx)
		},
	})
}

// AutoStartRuntimes starts all projects that were running before shutdown
func AutoStartRuntimes(
	lc fx.Lifecycle,
	logger *slog.Logger,
	projectService *service.ProjectService,
	pipelineService *service.PipelineService,
	runtimeManager *runtime.Manager,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting autostart process...")

				// Find all projects with status "running"
				projects, err := projectService.GetByStatus(ctx, domain.ProjectStatusRunning)
				if err != nil {
					logger.Error("Failed to get running projects", "error", err)
					return
				}

				if len(projects) == 0 {
					logger.Info("No running projects to autostart")
					return
				}

				logger.Info("Found projects to autostart", "count", len(projects))

				for _, project := range projects {
					// Get active release code
					release, err := pipelineService.GetActiveRelease(ctx, project.ID)
					if err != nil {
						logger.Warn("Failed to get active release for project, skipping",
							"project", project.Slug, "error", err)
						// Set project status to stopped since we can't start it
						projectService.UpdateStatus(ctx, project.ID, domain.ProjectStatusStopped)
						continue
					}

					// Start runtime
					if err := runtimeManager.Start(ctx, project.ID, release.Code); err != nil {
						logger.Error("Failed to autostart project",
							"project", project.Slug, "error", err)
						projectService.UpdateStatus(ctx, project.ID, domain.ProjectStatusStopped)
						continue
					}

					logger.Info("Autostarted project", "project", project.Slug, "release", release.Version)
				}

				logger.Info("Autostart process completed")
			}()
			return nil
		},
	})
}

func New(configPath string) *fx.App {
	return fx.New(
		fx.Provide(
			func() (*config.Config, error) {
				return config.Load(configPath)
			},
			NewLogger,
			NewGinEngine,

			// Database
			repository.NewMongoDB,

			// Repositories
			repository.NewUserRepository,
			repository.NewProjectRepository,
			repository.NewGoalRepository,
			repository.NewPipelineRepository,
			repository.NewEnvironmentRepository,
			repository.NewModelRepository,
			repository.NewWidgetRepository,

			// Services
			service.NewAuthService,
			service.NewUserService,
			service.NewProjectService,
			service.NewGoalService,
			service.NewPipelineService,
			service.NewEnvironmentService,
			service.NewStorageService,
			service.NewModelService,
			service.NewWidgetService,

			// Runtime
			runtime.NewManager,
			plugin.NewLoader,

			// Middleware
			middleware.NewAuthMiddleware,

			// Handlers
			handler.NewAuthHandler,
			handler.NewUserHandler,
			handler.NewProjectHandler,
			handler.NewGoalHandler,
			handler.NewPipelineHandler,
			handler.NewStorageHandler,
			handler.NewModelHandler,
			handler.NewEnvironmentHandler,
			handler.NewRuntimeHandler,
			handler.NewWidgetHandler,
		),
		fx.Invoke(RegisterRoutes, StartServer, AutoStartRuntimes),
	)
}
