package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/levskiy0/m3m/internal/app"
	"github.com/levskiy0/m3m/internal/config"
	"github.com/levskiy0/m3m/internal/repository"
	"github.com/levskiy0/m3m/internal/runtime/modules"
	"github.com/levskiy0/m3m/internal/service"
	"github.com/levskiy0/m3m/pkg/schema"
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

var docsCmd = &cobra.Command{
	Use:   "docs [output-file]",
	Short: "Generate JavaScript API documentation",
	Long:  `Generate Markdown documentation for the M3M JavaScript runtime API. Default output: CODE.md`,
	Run: func(cmd *cobra.Command, args []string) {
		outputFile := "CODE.md"
		if len(args) > 0 {
			outputFile = args[0]
		}

		format, _ := cmd.Flags().GetString("format")

		// Get all module schemas
		schemas := modules.GetAllSchemas()

		var content string
		switch format {
		case "typescript", "ts":
			content = schema.GenerateAllTypeScript(schemas)
		case "markdown", "md":
			content = schema.GenerateAllMarkdown(schemas)
		default:
			content = schema.GenerateAllMarkdown(schemas)
		}

		// Write to file
		err := os.WriteFile(outputFile, []byte(content), 0644)
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Documentation generated: %s\n", outputFile)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "config file path")

	docsCmd.Flags().StringP("format", "f", "markdown", "Output format: markdown, typescript")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(newAdminCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(docsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
