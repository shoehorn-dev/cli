package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cfgFile       string
	profile       string
	noInteractive bool
	outputFormat  string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "shoehorn",
	Short: "Shoehorn CLI - Internal Developer Portal",
	Long: `Shoehorn CLI provides command-line access to the Shoehorn platform.

Use it to authenticate, manage workflows, and interact with the Forge service.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.shoehorn/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "default", "authentication profile to use")
	rootCmd.PersistentFlags().BoolVarP(&noInteractive, "no-interactive", "I", false, "disable interactive mode (force plain output)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "output format (table|json|yaml)")

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Shoehorn CLI v0.1.0")
		},
	})
}
