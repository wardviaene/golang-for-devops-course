package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wardviaene/golang-for-devops-course/tls-demo/pkg/key"
)

var keyPath string

func init() {
	rootCmd.AddCommand(keyCmd)
	rootCmd.Flags().StringVarP(&keyPath, "out", "o", "", "destination path")
}

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "key commands",
	Long:  `Commands to manage (RSA) keys`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Not enough commands given. Possible sub-commands for key: create")
			return
		}
		if keyPath == "" {
			keyPath = "key.pem"
		}
		if err := key.CreateRSAPrivateKeyAndSave(keyPath, 4096); err != nil {
			fmt.Printf("Create key error: %s\n", err)
			return
		}
		fmt.Printf("Created RSA Key: %s\n", keyPath)
	},
}
