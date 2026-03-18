package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/the-maldridge/potd/pkg/password"
)

var (
	decodeCmd = &cobra.Command{
		Use:   "decode",
		Short: "decode a random password",
		Long:  decodeCmdLongDocs,
		Run:   decodeCmdRun,
	}

	decodeCmdLongDocs = `potd decode <hostname> <token>

The decode command prooduces a password matching a given hostname and
ephemeral token.  This uses the same process as the password
generation to statelessly generate the same password as was previously
generated.

Several password modes are supported and can be specified by their
mode number:

  1 - Random Hexadecimal String
`
	decodeCmdPasswdSize      int
	decodeCmdPasswdMode      uint8
	decodeCmdSharedTokenFile string
)

func init() {
	decodeCmd.Flags().IntVarP(&decodeCmdPasswdSize, "size", "s", 24, "Size of the password to decode")
	decodeCmd.Flags().Uint8VarP(&decodeCmdPasswdMode, "mode", "m", 1, "Mode of password generation")
	decodeCmd.Flags().StringVarP(&decodeCmdSharedTokenFile, "shared-token", "t", "/usr/share/potd/shared_token", "Shared token file location")
	rootCmd.AddCommand(decodeCmd)
}

func decodeCmdRun(c *cobra.Command, args []string) {
	content, err := os.ReadFile(decodeCmdSharedTokenFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading shared token file: %s\n", err)
		os.Exit(1)
	}
	components := []string{args[0], args[1], string(content)}

	p := password.New(components, password.Mode(decodeCmdPasswdMode), decodeCmdPasswdSize)
	fmt.Println(p)
}
