/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"gogs.tail02d447.ts.net/rafael/service-cli/internal/util"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new service",
	Long:  `Create a new service with specified configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		img, _ := cmd.Flags().GetString("img")
		port, _ := cmd.Flags().GetString("port")
		bare, _ := cmd.Flags().GetBool("bare")

		if bare {
			// Create service with not TS key and serve file
			serviceDir, err := util.GenerateFolderService(name)
			if err != nil {
				fmt.Printf("Error creating service directory: %v\n", err)
				return
			}
			if err := util.GenerateComposeFile(serviceDir, name, img, bare); err != nil {
				fmt.Printf("Error creating compose file: %v\n", err)
				return
			}
			fmt.Printf("✓ Bare service '%s' created successfully\n", name)
			return
		}

		// Create service folder in the services directory
		serviceDir, err := util.GenerateFolderService(name)
		if err != nil {
			fmt.Printf("Error creating service directory: %v\n", err)
			return
		}

		// Handle file generations
		var wg sync.WaitGroup
		errCh := make(chan error, 3)

		wg.Add(3)
		go func() {
			defer wg.Done()
			if err := util.GenerateComposeFile(serviceDir, name, img, bare); err != nil {
				errCh <- fmt.Errorf("compose file: %w", err)
			}
		}()
		go func() {
			defer wg.Done()
			if err := util.GenerateTsKey(serviceDir, name); err != nil {
				errCh <- fmt.Errorf("ts key: %w", err)
			}
		}()
		go func() {
			defer wg.Done()
			if err := util.GenerateServeFile(serviceDir, port); err != nil {
				errCh <- fmt.Errorf("serve file: %w", err)
			}
		}()

		wg.Wait()
		close(errCh)

		var hasErrors bool
		for err := range errCh {
			fmt.Printf("Error: %v\n", err)
			hasErrors = true
		}
		if hasErrors {
			os.Exit(1)
		}

		fmt.Printf("✓ Service '%s' created successfully\n", name)
	},
}

func init() {
	createCmd.Flags().StringP("name", "n", "", "Name of the service to create")
	createCmd.Flags().StringP("img", "i", "", "Docker image for the service")
	createCmd.Flags().StringP("port", "p", "8080", "Port to expose the service on")
	createCmd.Flags().BoolP("bare", "b", false, "Create a bare service without additional files")
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("img")
	rootCmd.AddCommand(createCmd)
}
