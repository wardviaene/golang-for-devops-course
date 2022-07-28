package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create CA, certs, or keys",
	Long:  `commands to create resources (ca, certs, keys)`,
}
