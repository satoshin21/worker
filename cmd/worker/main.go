package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/satoshin21/worker/internal/command"
)

func main() {
	root := &cobra.Command{
		Use:           "worker",
		Short:         "Create git worktrees and open them in zellij tabs",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(command.NewCreateCommand())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
