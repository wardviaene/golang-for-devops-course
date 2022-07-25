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
	certCmd.Flags().StringVarP(&caCertPath, "crtOut", "c", "", "destination path for cert")
	certCmd.Flags().StringVarP(&caKeyPath, "keyOut", "k", "", "destination path for cert key")
	certCmd.Flags().StringVar(&caKeyPath, "caKey", "", "caKey to sign certificate")
	certCmd.Flags().StringVar(&caCertPath, "caCert", "", "CA Certificate")
	certCmd.Flags().StringVar(&certName, "name", "", "name of certificate (need to match the name in the config)")
	certCmd.MarkFlagRequired("caKey")
	certCmd.MarkFlagRequired("caCert")
	rootCmd.AddCommand(certCmd)
}

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Certificate commands",
	Long:  `Commands to manage a certificate`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Not enough commands given. Possible sub-commands for cert: create")
			return
		}
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
