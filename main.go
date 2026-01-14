/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/joho/godotenv"
	"gogs.tail02d447.ts.net/rafael/service-cli/cmd"
)

func main() {
	godotenv.Load()
	cmd.Execute()
}
