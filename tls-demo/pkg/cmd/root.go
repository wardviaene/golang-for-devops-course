package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wardviaene/golang-for-devops-course/tls-demo/pkg/cert"
	"gopkg.in/yaml.v2"
)

type Config struct {
	CACert *cert.CACert          `yaml:"caCert"`
	Cert   map[string]*cert.Cert `yaml:"certs"`
}

var (
	cfgFile       string
	cfgFileParsed Config
	rootCmd       = &cobra.Command{
		Use:   "tls",
		Short: "tls is a command line tool for TLS",
		Long: `tls is a command line tool for TLS.
			Mainly used for generation of X.509 certificates, but can be extended`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "tls.yaml", "config file")
}

func initConfig() {
	if cfgFile != "" {
		cfgFileBytes, err := os.ReadFile(cfgFile)
		if err != nil {
			fmt.Printf("Config Read error: %s\n", err)
			return
		}
		err = yaml.Unmarshal(cfgFileBytes, &cfgFileParsed)
		if err != nil {
			fmt.Printf("Config Parse error: %s\n", err)
			return
		}
	}
}
