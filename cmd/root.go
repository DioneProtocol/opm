// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DioneProtocol/odysseygo/utils/wrappers"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/DioneProtocol/opm/config"
	"github.com/DioneProtocol/opm/constant"
	"github.com/DioneProtocol/opm/opm"
)

var (
	goPath  = os.ExpandEnv("$GOPATH")
	homeDir = os.ExpandEnv("$HOME")
	opmDir  = filepath.Join(homeDir, fmt.Sprintf(".%s", constant.AppName))
)

const (
	configFileKey       = "config-file"
	opmPathKey          = "opm-path"
	pluginPathKey       = "plugin-path"
	credentialsFileKey  = "credentials-file"
	adminAPIEndpointKey = "admin-api-endpoint"
)

func New(fs afero.Fs) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "opm",
		Short: "opm is a plugin manager to help manage virtual machines and subnets",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// we need to initialize our config here before each command starts,
			// since Cobra doesn't actually parse any of the flags until
			// cobra.Execute() is called.
			return initializeConfig()
		},
	}

	rootCmd.PersistentFlags().String(configFileKey, "", "path to configuration file for the opm")
	rootCmd.PersistentFlags().String(opmPathKey, opmDir, "path to the directory opm creates its artifacts")
	rootCmd.PersistentFlags().String(pluginPathKey, filepath.Join(goPath, "src", "github.com", "DioneProtocol", "odysseygo", "build", "plugins"), "path to odyssey plugin directory")
	rootCmd.PersistentFlags().String(credentialsFileKey, "", "path to credentials file")
	rootCmd.PersistentFlags().String(adminAPIEndpointKey, "127.0.0.1:9650/ext/admin", "endpoint for the odyssey admin api")

	errs := wrappers.Errs{}
	errs.Add(
		viper.BindPFlag(configFileKey, rootCmd.PersistentFlags().Lookup(configFileKey)),
		viper.BindPFlag(opmPathKey, rootCmd.PersistentFlags().Lookup(opmPathKey)),
		viper.BindPFlag(pluginPathKey, rootCmd.PersistentFlags().Lookup(pluginPathKey)),
		viper.BindPFlag(credentialsFileKey, rootCmd.PersistentFlags().Lookup(credentialsFileKey)),
		viper.BindPFlag(adminAPIEndpointKey, rootCmd.PersistentFlags().Lookup(adminAPIEndpointKey)),
	)
	if errs.Errored() {
		return nil, errs.Err
	}

	rootCmd.AddCommand(
		install(fs),
		uninstall(fs),
		update(fs),
		upgrade(fs),
		listRepositories(fs),
		joinSubnet(fs),
		addRepository(fs),
		removeRepository(fs),
	)

	return rootCmd, nil
}

// initializes config from file, if available.
func initializeConfig() error {
	if viper.IsSet(configFileKey) {
		cfgFile := os.ExpandEnv(viper.GetString(configFileKey))
		viper.SetConfigFile(cfgFile)

		return viper.ReadInConfig()
	}

	return nil
}

// If we need to use custom git credentials (say for private repos).
// the zero value for credentials is safe to use.
func initCredentials() (http.BasicAuth, error) {
	result := http.BasicAuth{}

	if viper.IsSet(credentialsFileKey) {
		credentials := &config.Credential{}

		bytes, err := os.ReadFile(viper.GetString(credentialsFileKey))
		if err != nil {
			return result, err
		}
		if err := yaml.Unmarshal(bytes, credentials); err != nil {
			return result, err
		}

		result.Username = credentials.Username
		result.Password = credentials.Password
	}

	return result, nil
}

func initOPM(fs afero.Fs) (*opm.OPM, error) {
	credentials, err := initCredentials()
	if err != nil {
		return nil, err
	}

	return opm.New(opm.Config{
		Directory:        viper.GetString(opmPathKey),
		Auth:             credentials,
		AdminAPIEndpoint: viper.GetString(adminAPIEndpointKey),
		PluginDir:        viper.GetString(pluginPathKey),
		Fs:               fs,
	})
}
