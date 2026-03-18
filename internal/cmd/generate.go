package cmd

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/google/renameio/v2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/the-maldridge/potd/pkg/password"
)

var (
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "generate a random password, optionally applying it",
		Long:  generateCmdLongDocs,
		Run:   generateCmdRun,
	}

	generateCmdLongDocs = `potd generate

The generate command generates a random password string and can
optionally apply it to a shadow file (by default /etc/shadow).

Several password modes are supported and can be specified by their
mode number:

  1 - Random Hexadecimal String
`
	generateCmdPasswdSize      int
	generateCmdPasswdMode      uint8
	generateCmdModifyShadow    bool
	generateCmdShadowFile      string
	generateCmdShadowUser      string
	generateCmdSharedTokenFile string
	generateCmdIssueTemplate   string
	generateCmdIssueFile       string
)

func init() {
	generateCmd.Flags().IntVarP(&generateCmdPasswdSize, "size", "s", 24, "Size of the password to generate")
	generateCmd.Flags().Uint8VarP(&generateCmdPasswdMode, "mode", "m", 1, "Mode of password generation")
	generateCmd.Flags().BoolVarP(&generateCmdModifyShadow, "apply-shadow", "w", false, "Modify the shadow file")
	generateCmd.Flags().StringVarP(&generateCmdShadowFile, "shadow", "f", "/etc/shadow", "Shadow file location")
	generateCmd.Flags().StringVarP(&generateCmdSharedTokenFile, "shared-token", "t", "/usr/share/potd/shared_token", "Shared token file location")
	generateCmd.Flags().StringVarP(&generateCmdShadowUser, "user", "u", "root", "User to update")
	generateCmd.Flags().StringVarP(&generateCmdIssueTemplate, "issue-tpl", "", "/etc/issue.tpl", "Issue file template")
	generateCmd.Flags().StringVarP(&generateCmdIssueFile, "issue", "", "/etc/issue", "Issue file")
	rootCmd.AddCommand(generateCmd)
}

func generateCmdRun(c *cobra.Command, args []string) {
	name, _ := os.Hostname()
	token := uuid.New().String()

	content, err := os.ReadFile(generateCmdSharedTokenFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading shared token file: %s\n", err)
		os.Exit(1)
	}
	components := []string{name, token, string(content)}

	p := password.New(components, password.Mode(generateCmdPasswdMode), generateCmdPasswdSize)
	fmt.Println("Password: ", p)
	fmt.Println("Token: ", token)
	if !generateCmdModifyShadow {
		return
	}
	if err := p.UpdateShadow(generateCmdShadowFile, generateCmdShadowUser); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating shadow: %s\n", err)
		os.Exit(1)
	}

	tpl, err := template.ParseFiles(generateCmdIssueTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Not updating %s; Error parsing the issue template: %s\n", generateCmdIssueFile, err)
		return
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, token); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %s\n", err)
		os.Exit(1)
	}
	if err := renameio.WriteFile(generateCmdIssueFile, buf.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", generateCmdIssueFile, err)
		os.Exit(1)
	}
	fmt.Println("Update Complete")
}
