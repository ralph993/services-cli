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

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an existing service",
	Long:  "Delete an existing service by specifying its name.",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		removeKey, _ := cmd.Flags().GetBool("remove-key")

		// Handle file deletions
		if removeKey {
			var wg sync.WaitGroup
			errCh := make(chan error, 2)

			wg.Add(2)
			go func() {
				defer wg.Done()
				err := util.RemoveServiceFolder(name)
				if err != nil {
					errCh <- fmt.Errorf("service folder: %w", err)
				}
			}()
			go func() {
				defer wg.Done()
				err := util.RevokeTsKey(name)
				if err != nil {
					errCh <- fmt.Errorf("tailscale key: %w", err)
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
		} else {
			err := util.RemoveServiceFolder(name)
			if err != nil {
				fmt.Printf("Error removing service folder: %v\n", err)
				return
			}
		}

		fmt.Printf("Service '%s' deleted successfully.\n", name)
	},
}

func init() {
	deleteCmd.Flags().StringP("name", "n", "", "Name of the service to delete")
	deleteCmd.Flags().BoolP("remove-key", "r", false, "Also remove the Tailscale key associated with the service")
	deleteCmd.MarkFlagRequired("name")
	rootCmd.AddCommand(deleteCmd)
}
