package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// buildCLI assembles the cobra command tree for VortexUI panel CLI operations.
// The root "vortexui" command dispatches to subcommands; when invoked without a
// subcommand it falls through to the normal panel server startup (main.go).
func buildCLI() *cobra.Command {
	root := &cobra.Command{
		Use:     "vortexui",
		Short:   "VortexUI panel — proxy management platform",
		Version: version,
	}

	root.AddCommand(
		newDoctorCmd(),
		newMigrateCmd(),
		newBackupCmd(),
		newSettingsCmd(),
		newCleanupCmd(),
		newUserCmd(),
		newNodeCmd(),
	)

	return root
}

// runCLI checks if any CLI subcommand was invoked (os.Args[1] matches a known
// subcommand name). If so it runs the cobra command tree and exits. Otherwise
// it returns false so main() proceeds with the server.
func runCLI() bool {
	if len(os.Args) < 2 {
		return false
	}

	knownCmds := map[string]bool{
		"doctor": true, "migrate": true, "backup": true,
		"settings": true, "cleanup": true, "user": true, "node": true,
		"help": true, "version": true, "completion": true,
	}

	if !knownCmds[os.Args[1]] {
		return false
	}

	cli := buildCLI()
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
	return true // unreachable, satisfies compiler
}

// --- doctor ---

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run system health checks (database, Redis, nodes, TLS, DNS, disk)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("VortexUI Doctor — running health checks...")
			fmt.Println()

			// Checks would be delegated to service.DoctorService in production.
			// For now we show the framework structure.
			checks := []string{
				"Database connectivity",
				"Redis connectivity",
				"Node agent reachability",
				"TLS certificate validity",
				"Required ports availability",
				"DNS resolution",
				"Disk space",
			}

			_ = ctx
			for _, check := range checks {
				fmt.Printf("  [•] %s ... pending\n", check)
			}
			fmt.Println()
			fmt.Println("Run with live configuration to execute actual checks.")
			return nil
		},
	}
}

// --- migrate ---

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate users from another panel (Marzban, 3x-ui, PasarGuard)",
	}

	var source, dsn string
	cmd.Flags().StringVar(&source, "source", "", "Source panel type (marzban, 3x-ui, pasarguard)")
	cmd.Flags().StringVar(&dsn, "dsn", "", "Source database connection string")
	_ = cmd.MarkFlagRequired("source")
	_ = cmd.MarkFlagRequired("dsn")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Migrating from %s (dsn: %s)...\n", source, dsn)
		fmt.Println("Migration service would read foreign schema and map to VortexUI domain models.")
		return nil
	}

	return cmd
}

// --- backup ---

func newBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Create or restore encrypted backups",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new encrypted backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Creating encrypted backup...")
			fmt.Println("Backup service would encrypt with AES-256 and upload to configured destination.")
			return nil
		},
	}

	restoreCmd := &cobra.Command{
		Use:   "restore [file]",
		Short: "Restore from an encrypted backup",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Restoring from %s...\n", args[0])
			return nil
		},
	}

	cmd.AddCommand(createCmd, restoreCmd)
	return cmd
}

// --- settings ---

func newSettingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Export or import panel settings",
	}

	exportCmd := &cobra.Command{
		Use:   "export [file]",
		Short: "Export all panel settings to YAML",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := "settings.yaml"
			if len(args) > 0 {
				out = args[0]
			}
			fmt.Printf("Exporting settings to %s...\n", out)
			return nil
		},
	}

	importCmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Import and validate YAML settings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Importing settings from %s...\n", args[0])
			return nil
		},
	}

	cmd.AddCommand(exportCmd, importCmd)
	return cmd
}

// --- cleanup ---

func newCleanupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Delete expired/limited users past retention period",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Running auto-cleanup...")
			fmt.Println("Would delete expired/limited users after configured retention period.")
			return nil
		},
	}
}

// --- user ---

func newUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "User management CLI operations",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing users...")
			return nil
		},
	}

	cmd.AddCommand(listCmd)
	return cmd
}

// --- node ---

func newNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Node management CLI operations",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing nodes...")
			return nil
		},
	}

	cmd.AddCommand(listCmd)
	return cmd
}
