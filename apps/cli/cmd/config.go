package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Taichi-iskw/yt-lang/internal/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  `Manage configuration settings for yt-lang.`,
}

// configInitCmd represents the config init command
var configInitCmd = &cobra.Command{
	Use:   "init [DATABASE_URL]",
	Short: "Initialize configuration file",
	Long:  `Create a new configuration file with database connection settings.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var databaseURL string
		if len(args) > 0 {
			databaseURL = args[0]
		}

		if err := config.InitConfig(databaseURL); err != nil {
			return err
		}

		configPath, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		fmt.Printf("Created configuration file: %s\n", configPath)
		fmt.Println("Please edit the database_url in this file to match your PostgreSQL database.")

		return nil
	},
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration file path and settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		fmt.Printf("Configuration file: %s\n\n", configPath)

		// Load and display current config
		cfg, err := config.NewConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		fmt.Printf("DATABASE_URL: %s\n", cfg.DatabaseURL)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
}
