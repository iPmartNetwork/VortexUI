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
	if len(args) == 0 {
		return fmt.Errorf("usage: panel admin create|reset-password")
	}
	switch args[0] {
	case "create":
		return runAdminCreate(ctx, args[1:])
	case "reset-password":
		return runAdminResetPassword(ctx, args[1:])
	default:
		return fmt.Errorf("usage: panel admin create|reset-password")
	}
}

func runAdminCreate(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("admin create", flag.ContinueOnError)
	username := fs.String("username", "", "admin username")
	password := fs.String("password", "", "admin password")
	sudo := fs.Bool("sudo", true, "grant full (sudo) privileges")
	enableTOTP := fs.Bool("totp", false, "enable TOTP 2FA and print enrollment URL")
	if err := fs.Parse(args); err != nil {
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

	svc := service.NewAdminService(store.Admins(), store.Users())
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

func runAdminResetPassword(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("admin reset-password", flag.ContinueOnError)
	username := fs.String("username", "", "admin username")
	password := fs.String("password", "", "new admin password")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *username == "" || *password == "" {
		return fmt.Errorf("usage: panel admin reset-password --username U --password P")
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

	svc := service.NewAdminService(store.Admins(), store.Users())
	if err := svc.ResetPassword(ctx, *username, *password); err != nil {
		return err
	}
	fmt.Printf("password reset for admin %q\n", *username)
	return nil
}
