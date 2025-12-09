package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/levskiy0/m3m/internal/app"
	"github.com/levskiy0/m3m/internal/config"
	"github.com/levskiy0/m3m/internal/repository"
	"github.com/levskiy0/m3m/internal/service"
)

var configFile string

var rootCmd = &cobra.Command{
	Use:   "m3m",
	Short: "M3M - Mini Services Manager",
	Long:  `M3M is a platform for creating and managing mini-services/workers in JavaScript`,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the M3M server",
	Run: func(cmd *cobra.Command, args []string) {
		application := app.New(configFile)
		application.Run()
	},
}

var newAdminCmd = &cobra.Command{
	Use:   "new-admin [email] [password]",
	Short: "Create a new admin user",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		password := args[1]

		// Load config
		cfg, err := config.Load(configFile)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Connect to MongoDB
		db, err := repository.NewMongoDB(cfg)
		if err != nil {
			fmt.Printf("Error connecting to MongoDB: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Create repositories and services
		userRepo := repository.NewUserRepository(db)
		authService := service.NewAuthService(userRepo, cfg)
		userService := service.NewUserService(userRepo, authService)

		// Create root user
		ctx := context.Background()
		user, err := userService.CreateRootUser(ctx, email, password)
		if err != nil {
			fmt.Printf("Error creating admin: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Root admin created successfully!\n")
		fmt.Printf("Email: %s\n", user.Email)
		fmt.Printf("ID: %s\n", user.ID.Hex())
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("M3M %s\n", app.Version)
	},
}

var pluginBuildCmd = &cobra.Command{
	Use:   "plugin-build [plugin-name]",
	Short: "Build a plugin from /app/data/plugins/{plugin-name}",
	Long: `Build a Go plugin from source located in /app/data/plugins/{plugin-name}.
The plugin must have a plugin.go file with NewPlugin() function.
Output .so file will be placed in /app/plugins/`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pluginName := args[0]

		// Paths for Docker environment
		srcDir := filepath.Join("/app/data/plugins", pluginName)
		outputDir := "/app/plugins"
		outputFile := filepath.Join(outputDir, pluginName+".so")

		// Check if source directory exists
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			fmt.Printf("Error: plugin source directory not found: %s\n", srcDir)
			os.Exit(1)
		}

		// Check if plugin.go exists
		pluginFile := filepath.Join(srcDir, "plugin.go")
		if _, err := os.Stat(pluginFile); os.IsNotExist(err) {
			fmt.Printf("Error: plugin.go not found in %s\n", srcDir)
			os.Exit(1)
		}

		// Ensure output directory exists
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Building plugin: %s\n", pluginName)
		fmt.Printf("Source: %s\n", srcDir)
		fmt.Printf("Output: %s\n", outputFile)

		// Run go build with plugin mode
		buildCmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outputFile, ".")
		buildCmd.Dir = srcDir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		buildCmd.Env = append(os.Environ(),
			"CGO_ENABLED=1",
			"GOOS=linux",
		)

		if err := buildCmd.Run(); err != nil {
			fmt.Printf("Error building plugin: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Plugin %s built successfully!\n", pluginName)
		fmt.Printf("Restart the server to load the new plugin.\n")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "config file path")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(newAdminCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pluginBuildCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
