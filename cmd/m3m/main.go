package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"m3m/internal/app"
	"m3m/internal/config"
	"m3m/internal/repository"
	"m3m/internal/service"
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
		fmt.Println("M3M v1.0.0")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "config file path")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(newAdminCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
