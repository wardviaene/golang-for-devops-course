package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create ca, cert or keys",
	Long:  `Commands to create resources (ca, cert, keys)`,
}
