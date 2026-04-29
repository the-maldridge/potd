package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/google/renameio/v2"
	"github.com/spf13/cobra"

	"github.com/the-maldridge/potd/pkg/password"
	"github.com/the-maldridge/potd/pkg/types"
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
  2 - XKCD Style Multi-Word String
`
	generateCmdPasswdSize    int
	generateCmdPasswdMode    uint8
	generateCmdModifyShadow  bool
	generateCmdShadowFile    string
	generateCmdShadowUser    string
	generateCmdIssueTemplate string
	generateCmdIssueFile     string
	generateCmdClientCert    string
	generateCmdClientKey     string
	generateCmdResolverHost  string
	generateCmdInsecureTLS   bool
)

func init() {
	generateCmd.Flags().IntVarP(&generateCmdPasswdSize, "size", "s", 5, "Size of the password to generate")
	generateCmd.Flags().Uint8VarP(&generateCmdPasswdMode, "mode", "m", 2, "Mode of password generation")
	generateCmd.Flags().BoolVarP(&generateCmdModifyShadow, "apply-shadow", "w", false, "Modify the shadow file")
	generateCmd.Flags().StringVarP(&generateCmdShadowFile, "shadow", "f", "/etc/shadow", "Shadow file location")
	generateCmd.Flags().StringVarP(&generateCmdShadowUser, "user", "u", "root", "User to update")
	generateCmd.Flags().StringVarP(&generateCmdIssueTemplate, "issue-tpl", "", "/etc/issue.tpl", "Issue file template")
	generateCmd.Flags().StringVarP(&generateCmdIssueFile, "issue", "", "/etc/issue", "Issue file")
	generateCmd.Flags().StringVarP(&generateCmdClientCert, "cert", "", "", "Client Certificate")
	generateCmd.Flags().StringVarP(&generateCmdClientKey, "key", "", "", "Client Certificate Key")
	generateCmd.Flags().StringVarP(&generateCmdResolverHost, "resolver", "", "", "Resolver instance to update escrowed tokens")
	generateCmd.Flags().BoolVarP(&generateCmdInsecureTLS, "insecure-tls", "", false, "Skip TLS verification (INSECURE)")
	rootCmd.AddCommand(generateCmd)
}

func generateCmdRun(c *cobra.Command, args []string) {
	name, _ := os.Hostname()
	e := types.EscrowedToken{
		Host:  name,
		Token: password.ChallengeToken(5),
	}

	challenge := password.ChallengeToken(4)

	components := []string{e.Host, challenge, e.Token}

	p := password.New(components, password.Mode(generateCmdPasswdMode), generateCmdPasswdSize)
	fmt.Println("Password: ", p)
	fmt.Println("Token: ", challenge)

	cert, err := tls.LoadX509KeyPair(generateCmdClientCert, generateCmdClientKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading client certificate: %s\n", err)
		os.Exit(2)
	}
	h := http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: generateCmdInsecureTLS,
			},
		},
	}

	// This can't error since its a fixed type
	b, _ := json.Marshal(e)
	bf := bytes.NewBuffer(b)
	u := url.URL{
		Scheme: "https",
		Host:   generateCmdResolverHost,
		Path:   "/api/escrow/update-token",
	}
	req, _ := http.NewRequest("POST", u.String(), bf)
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error escrowing token: %s\n", err)
		if resp != nil {
			fmt.Fprintf(os.Stderr, "Code: %d\n", resp.StatusCode)
		}
		os.Exit(1)
	}

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
	if err := tpl.Execute(&buf, challenge); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %s\n", err)
		os.Exit(1)
	}
	if err := renameio.WriteFile(generateCmdIssueFile, buf.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", generateCmdIssueFile, err)
		os.Exit(1)
	}
	fmt.Println("Update Complete")
}
