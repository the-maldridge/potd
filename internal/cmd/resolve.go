package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/the-maldridge/potd/pkg/password"
)

var (
	resolveCmd = &cobra.Command{
		Use:   "resolve",
		Short: "resolve a random password",
		Long:  resolveCmdLongDocs,
		Run:   resolveCmdRun,
	}

	resolveCmdLongDocs = `potd resolve <hostname> <token>

The resolve command prooduces a password matching a given hostname and
ephemeral token.  This uses the same process as the password
generation to statelessly generate the same password as was previously
generated.

Several password modes are supported and can be specified by their
mode number:

  1 - Random Hexadecimal String
`
	resolveCmdPasswdSize      int
	resolveCmdPasswdMode      uint8
	resolveCmdSharedTokenFile string
)

func init() {
	resolveCmd.Flags().IntVarP(&resolveCmdPasswdSize, "size", "s", 5, "Size of the password to resolve")
	resolveCmd.Flags().Uint8VarP(&resolveCmdPasswdMode, "mode", "m", 2, "Mode of password generation")
	resolveCmd.Flags().StringVarP(&resolveCmdSharedTokenFile, "shared-token", "t", "/usr/share/potd/shared_token", "Shared token file location")
	rootCmd.AddCommand(resolveCmd)
}

func resolveCmdRun(c *cobra.Command, args []string) {
	content, err := os.ReadFile(resolveCmdSharedTokenFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading shared token file: %s\n", err)
		os.Exit(1)
	}
	components := []string{args[0], args[1], string(content)}

	p := password.New(components, password.Mode(resolveCmdPasswdMode), resolveCmdPasswdSize)
	fmt.Println(p)
}
