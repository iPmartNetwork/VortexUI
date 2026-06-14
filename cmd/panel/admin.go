package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/vortexui/vortexui/internal/config"
	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/platform/postgres"
)

// runAdmin implements the `panel admin <sub>` command group. It is the bootstrap
// path: without it no one can ever log in, since the API never self-seeds an
// admin (auto-seeding a default credential would be a security footgun).
func runAdmin(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "create" {
		return fmt.Errorf("usage: panel admin create --username U --password P [--sudo] [--totp]")
	}

	fs := flag.NewFlagSet("admin create", flag.ContinueOnError)
	username := fs.String("username", "", "admin username")
	password := fs.String("password", "", "admin password")
	sudo := fs.Bool("sudo", true, "grant full (sudo) privileges")
	enableTOTP := fs.Bool("totp", false, "enable TOTP 2FA and print enrollment URL")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	cfg, err := config.LoadPanel()
	if err != nil {
		return err
	}
	store, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer store.Close()

	svc := service.NewAdminService(store.Admins())
	admin, totpURL, err := svc.Create(ctx, service.CreateAdminInput{
		Username:   *username,
		Password:   *password,
		Sudo:       *sudo,
		EnableTOTP: *enableTOTP,
	})
	if err != nil {
		return err
	}

	fmt.Printf("created admin %q (id=%s sudo=%t)\n", admin.Username, admin.ID, admin.Sudo)
	if totpURL != "" {
		fmt.Printf("scan this TOTP enrollment URL in your authenticator (shown once):\n%s\n", totpURL)
	}
	return nil
}
