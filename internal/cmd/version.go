package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run:   versionCmdRun,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func versionCmdRun(c *cobra.Command, args []string) {
	fmt.Printf("Version: %s\n", Version)
}
