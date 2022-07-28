package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/wardviaene/golang-for-devops-course/tls-demo/pkg/cert"
	"gopkg.in/yaml.v2"
)

type Config struct {
	CACert *cert.CACert          `yaml:"caCert"`
	Cert   map[string]*cert.Cert `yaml:"certs"`
}

var cfgFilePath string
var config Config

var rootCmd = &cobra.Command{
	Use:   "tls",
	Short: "tls is a command line tool for TLS.",
	Long: `tls is a command line tool for TLS.
		Mainly used for generation of X.509 certificates, but can be extended`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFilePath, "config", "c", "", "config file (default is tls.yaml)")
}

func initConfig() {
	if cfgFilePath == "" {
		cfgFilePath = "tls.yaml"
	}
	cfgFileBytes, err := ioutil.ReadFile(cfgFilePath)
	if err != nil {
		fmt.Printf("Error while reading config file: %s\n", err)
		return
	}
	err = yaml.Unmarshal(cfgFileBytes, &config)
	if err != nil {
		fmt.Printf("Error while parsing config file: %s\n", err)
		return
	}
}
