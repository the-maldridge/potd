package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "potd",
		Short: "potd handles passwords of the day",
		Long:  potdLongDocs,
	}

	potdLongDocs = `potd - Password of the Day

Passwords that need to be rotated often aren't, and potd can help.  This tool both generates a password and can optionally edit Linux shadow files, and seperately provides a secure means of recovering passwords that have been generated.`
)
