package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wardviaene/golang-for-devops-course/tls-demo/pkg/cert"
)

var caCertPath string
var caKeyPath string

func init() {
	rootCmd.AddCommand(caCmd)
	rootCmd.Flags().StringVarP(&caCertPath, "crtOut", "c", "", "destination path for CA cert")
	rootCmd.Flags().StringVarP(&caKeyPath, "keyOut", "k", "", "destination path for CA key")
}

var caCmd = &cobra.Command{
	Use:   "ca",
	Short: "CA commands",
	Long:  `Commands to manage a Certificate Authority (CA)`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Not enough commands given. Possible sub-commands for ca: create")
			return
		}
		if caCertPath == "" {
			caCertPath = "ca.crt"
		}
		if caKeyPath == "" {
			caKeyPath = "ca.key"
		}

		if err := cert.CreateCACert(cfgFileParsed.CACert, caKeyPath, caCertPath); err != nil {
			fmt.Printf("Create ca error: %s\n", err)
			return
		}
		fmt.Printf("Created CA Key and Cert: %s and %s\n", caKeyPath, caCertPath)
	},
}
