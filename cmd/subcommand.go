package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ava-labs/apm/apm"
)

var (
	vmAlias     string
	subnetAlias string
)

func install() *cobra.Command {
	command := &cobra.Command{
		Use:   "install",
		Short: "installs a virtual machine by its alias",
	}
	command.PersistentFlags().StringVar(&vmAlias, "vm-alias", "", "vm alias to install")
	command.RunE = func(_ *cobra.Command, _ []string) error {
		apm, err := apm.New(apm.Config{
			Directory: apmDir,
			Auth:      authToken,
		})
		if err != nil {
			return err
		}
		return apm.Install(vmAlias)
	}

	return command
}

func listRepositories() *cobra.Command {
	command := &cobra.Command{
		Use:   "list-repositories",
		Short: "list registered source repositories for avalanche plugins",
	}
	command.RunE = func(_ *cobra.Command, _ []string) error {
		apm, err := apm.New(apm.Config{
			Directory: apmDir,
			Auth:      authToken,
		})
		if err != nil {
			return err
		}

		return apm.ListRepositories()
	}

	return command
}

func joinSubnet() *cobra.Command {
	command := &cobra.Command{
		Use:   "join",
		Short: "join a subnet by its alias.",
	}
	command.PersistentFlags().StringVar(&subnetAlias, "subnet-alias", "", "subnet alias to join")
	command.RunE = func(_ *cobra.Command, _ []string) error {
		apm, err := apm.New(apm.Config{
			Directory: apmDir,
			Auth:      authToken,
		})
		if err != nil {
			return err
		}

		return apm.JoinSubnet(subnetAlias)
	}

	return command
}