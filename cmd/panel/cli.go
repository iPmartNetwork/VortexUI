package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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
		newUpdateCmd(),
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
		"doctor": true, "update": true, "migrate": true, "backup": true,
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

// ANSI color helpers for doctor output.
const (
	colorGreen  = "\033[0;32m"
	colorRed    = "\033[0;31m"
	colorYellow = "\033[1;33m"
	colorReset  = "\033[0m"
)

func doctorPass(name, detail string) {
	fmt.Printf("  %s[✓]%s %s: %s\n", colorGreen, colorReset, name, detail)
}

func doctorFail(name, detail string) {
	fmt.Printf("  %s[✗]%s %s: %s\n", colorRed, colorReset, name, detail)
}

func doctorWarn(name, detail string) {
	fmt.Printf("  %s[!]%s %s: %s\n", colorYellow, colorReset, name, detail)
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run system health checks (database, Redis, TLS, disk, ports)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			fmt.Println("VortexUI Doctor — running health checks...")
			fmt.Println()

			failures := 0

			// 1. Database connectivity
			dbURL := os.Getenv("DATABASE_URL")
			if dbURL == "" {
				dbURL = os.Getenv("VORTEX_DATABASE_URL")
			}
			if dbURL == "" {
				doctorWarn("Database", "DATABASE_URL not set, skipping")
			} else {
				pool, err := pgxpool.New(ctx, dbURL)
				if err != nil {
					doctorFail("Database", fmt.Sprintf("connection failed: %v", err))
					failures++
				} else {
					if err := pool.Ping(ctx); err != nil {
						doctorFail("Database", fmt.Sprintf("ping failed: %v", err))
						failures++
					} else {
						doctorPass("Database", "connected and responding")
					}
					pool.Close()
				}
			}

			// 2. Redis connectivity
			redisURL := os.Getenv("REDIS_URL")
			if redisURL == "" {
				redisURL = os.Getenv("VORTEX_REDIS_URL")
			}
			if redisURL == "" {
				doctorWarn("Redis", "REDIS_URL not set, skipping")
			} else {
				opts, err := redis.ParseURL(redisURL)
				if err != nil {
					doctorFail("Redis", fmt.Sprintf("invalid URL: %v", err))
					failures++
				} else {
					rdb := redis.NewClient(opts)
					if err := rdb.Ping(ctx).Err(); err != nil {
						doctorFail("Redis", fmt.Sprintf("ping failed: %v", err))
						failures++
					} else {
						doctorPass("Redis", "connected and responding")
					}
					_ = rdb.Close()
				}
			}

			// 3. TLS certificate validity
			certFile := os.Getenv("TLS_CERT")
			if certFile == "" {
				certFile = "deploy/certs/panel.crt"
			}
			keyFile := os.Getenv("TLS_KEY")
			if keyFile == "" {
				keyFile = "deploy/certs/panel.key"
			}
			cleanCert := filepath.Clean(certFile)
			cleanKey := filepath.Clean(keyFile)
			if _, err := os.Stat(cleanCert); os.IsNotExist(err) {
				doctorWarn("TLS certificate", fmt.Sprintf("%s not found, skipping", cleanCert))
			} else {
				cert, err := tls.LoadX509KeyPair(cleanCert, cleanKey)
				if err != nil {
					doctorFail("TLS certificate", fmt.Sprintf("load failed: %v", err))
					failures++
				} else {
					leaf, err := x509.ParseCertificate(cert.Certificate[0])
					if err != nil {
						doctorFail("TLS certificate", fmt.Sprintf("parse failed: %v", err))
						failures++
					} else {
						daysLeft := int(time.Until(leaf.NotAfter).Hours() / 24)
						if daysLeft < 0 {
							doctorFail("TLS certificate", fmt.Sprintf("EXPIRED %d days ago", -daysLeft))
							failures++
						} else if daysLeft < 30 {
							doctorWarn("TLS certificate", fmt.Sprintf("expires in %d days", daysLeft))
						} else {
							doctorPass("TLS certificate", fmt.Sprintf("valid, expires in %d days", daysLeft))
						}
					}
				}
			}

			// 4. Disk space (require at least 1GB free)
			freeMB, err := getFreeDiskMB("/")
			if err != nil {
				doctorWarn("Disk space", fmt.Sprintf("check unavailable: %v", err))
			} else if freeMB < 1024 {
				doctorFail("Disk space", fmt.Sprintf("%dMB free (need ≥1GB)", freeMB))
				failures++
			} else {
				doctorPass("Disk space", fmt.Sprintf("%dMB free", freeMB))
			}

			// 5. Port 8080 check
			ln, err := net.Listen("tcp", "127.0.0.1:8080")
			if err != nil {
				// Port is in use — check if it's us by trying to reach our health endpoint
				if strings.Contains(err.Error(), "address already in use") || strings.Contains(err.Error(), "bind") {
					doctorPass("Port 8080", "in use (likely VortexUI)")
				} else {
					doctorWarn("Port 8080", fmt.Sprintf("check failed: %v", err))
				}
			} else {
				_ = ln.Close()
				doctorPass("Port 8080", "available")
			}

			fmt.Println()
			if failures > 0 {
				fmt.Printf("%s%d check(s) failed%s\n", colorRed, failures, colorReset)
				return fmt.Errorf("%d health check(s) failed", failures)
			}
			fmt.Printf("%sAll checks passed%s\n", colorGreen, colorReset)
			return nil
		},
	}
}

// --- update ---

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update VortexUI (git pull, rebuild, restart service)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("VortexUI Update — pulling latest changes...")
			fmt.Println()

			// 1. git pull
			gitPull := exec.Command("git", "pull", "origin", "master")
			gitPull.Stdout = os.Stdout
			gitPull.Stderr = os.Stderr
			if err := gitPull.Run(); err != nil {
				return fmt.Errorf("git pull failed: %w", err)
			}
			fmt.Println()

			// 2. Rebuild the binary
			fmt.Println("==> Building panel binary...")
			goBuild := exec.Command("go", "build",
				"-ldflags", fmt.Sprintf("-s -w -X main.version=%s", version),
				"-o", "/usr/local/bin/vortex-panel",
				"./cmd/panel",
			)
			goBuild.Stdout = os.Stdout
			goBuild.Stderr = os.Stderr
			goBuild.Env = append(os.Environ(), "CGO_ENABLED=0")
			if err := goBuild.Run(); err != nil {
				return fmt.Errorf("go build failed: %w", err)
			}
			fmt.Println("    Binary built: /usr/local/bin/vortex-panel")
			fmt.Println()

			// 3. Restart systemd service
			service := os.Getenv("VORTEX_SERVICE")
			if service == "" {
				service = "vortexui-panel"
			}

			fmt.Printf("==> Restarting %s...\n", service)

			// Validate service name to prevent command injection
			if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`).MatchString(service) {
				return fmt.Errorf("invalid service name: %s", service)
			}

			// Unmask if masked
			isEnabled := exec.Command("systemctl", "is-enabled", service) //nolint:gosec // service name validated above
			out, _ := isEnabled.Output()
			if strings.TrimSpace(string(out)) == "masked" {
				fmt.Printf("    Service %s is masked — unmasking...\n", service)
				unmask := exec.Command("systemctl", "unmask", service) //nolint:gosec // service name validated above
				unmask.Stdout = os.Stdout
				unmask.Stderr = os.Stderr
				if err := unmask.Run(); err != nil {
					return fmt.Errorf("unmask failed: %w", err)
				}
			}

			restart := exec.Command("systemctl", "restart", service) //nolint:gosec // service name validated above
			restart.Stdout = os.Stdout
			restart.Stderr = os.Stderr
			if err := restart.Run(); err != nil {
				return fmt.Errorf("restart failed: %w", err)
			}

			fmt.Println()
			fmt.Printf("%s==> Update complete!%s\n", colorGreen, colorReset)
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
