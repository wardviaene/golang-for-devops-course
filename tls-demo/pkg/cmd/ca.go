package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wardviaene/golang-for-devops-course/tls-demo/pkg/cert"
)

var caCertPath string
var caKeyPath string

func init() {
	createCmd.AddCommand(caCreateCmd)
	caCreateCmd.Flags().StringVarP(&caCertPath, "crt-out", "o", "", "destination path for CA cert")
	caCreateCmd.Flags().StringVarP(&caKeyPath, "key-out", "k", "", "destination path for CA key")
}

var caCreateCmd = &cobra.Command{
	Use:   "ca",
	Short: "CA commands",
	Long:  `Commands to manage a Certificate Authority (CA)`,
	Run: func(cmd *cobra.Command, args []string) {
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
