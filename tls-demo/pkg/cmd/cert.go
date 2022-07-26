package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/wardviaene/golang-for-devops-course/tls-demo/pkg/cert"
)

var certPath string
var certKeyPath string
var certName string

func init() {
	createCmd.AddCommand(certCreateCmd)
	certCreateCmd.Flags().StringVarP(&caCertPath, "crt-out", "o", "", "destination path for cert")
	certCreateCmd.Flags().StringVarP(&caKeyPath, "key-out", "k", "", "destination path for cert key")
	certCreateCmd.Flags().StringVar(&caKeyPath, "ca-key", "", "caKey to sign certificate")
	certCreateCmd.Flags().StringVar(&caCertPath, "ca-cert", "", "CA Certificate")
	certCreateCmd.Flags().StringVar(&certName, "cert-name", "", "name of certificate (need to match the name in the config)")
	certCreateCmd.MarkFlagRequired("caKey")
	certCreateCmd.MarkFlagRequired("caCert")
}

var certCreateCmd = &cobra.Command{
	Use:   "cert",
	Short: "Certificate commands",
	Long:  `Commands to manage a certificate`,
	Run: func(cmd *cobra.Command, args []string) {
		if certPath == "" {
			certPath = "server.crt"
		}
		if certKeyPath == "" {
			certKeyPath = "server.key"
		}

		caKeyBytes, err := ioutil.ReadFile(caKeyPath)
		if err != nil {
			fmt.Printf("CA Key Read error: %s", err)
			return
		}
		caCertBytes, err := ioutil.ReadFile(caCertPath)
		if err != nil {
			fmt.Printf("CA Cert Read error: %s", err)
			return
		}

		certSelected, ok := cfgFileParsed.Cert[certName]
		if !ok {
			fmt.Printf("Supply an existing cert name\n")
			return
		}
		if err := cert.CreateCert(certSelected, caKeyBytes, caCertBytes, certKeyPath, certPath); err != nil {
			fmt.Printf("Create ca error: %s\n", err)
			return
		}
		fmt.Printf("Created Server Key and Cert: %s and %s\n", caKeyPath, caCertPath)
	},
}
