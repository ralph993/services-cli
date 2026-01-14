/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gogs.tail02d447.ts.net/rafael/service-cli/internal/util"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all existing services",
	Long:  "List all existing services managed by the CLI tool.",
	Run: func(cmd *cobra.Command, args []string) {
		list, err := util.GetServiceList()
		if err != nil {
			fmt.Println("Error getting service list:", err)
			return
		}
		for _, service := range list {
			fmt.Println(service)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
