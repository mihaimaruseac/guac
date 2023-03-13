package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var flags = struct {
	addTestData    bool
	playgroundPort int
}{}

var rootCmd = &cobra.Command{
	Use:   "main",
	Short: "GraphQL playground for testing GUAC GraphQL backends and GraphQL integration",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

var cfgFile string

func init() {
	cmdFlags := rootCmd.Flags()
	cmdFlags.BoolVar(&flags.addTestData, "testdata", false, "Populate Neo4J database with test data")
	cmdFlags.IntVar(&flags.playgroundPort, "port", 8080, "Port to listen on for the GraphQL playground")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
