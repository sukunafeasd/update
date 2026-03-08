package config

import (
	"testing"
	"time"
)

func TestResolveBindAddressUsesPortForProduction(t *testing.T) {
	t.Setenv("UNIVERSALD_BIND", "")
	t.Setenv("PORT", "10000")

	if got := resolveBindAddress(); got != "0.0.0.0:10000" {
		t.Fatalf("expected production bind from PORT, got %q", got)
	}
}

func TestResolveBindAddressPrefersExplicitBind(t *testing.T) {
	t.Setenv("UNIVERSALD_BIND", "127.0.0.1:7788")
	t.Setenv("PORT", "10000")

	if got := resolveBindAddress(); got != "127.0.0.1:7788" {
		t.Fatalf("expected explicit bind to win, got %q", got)
	}
}

func TestDefaultOpenBrowserDisablesForHostedRuntime(t *testing.T) {
	t.Setenv("PORT", "10000")
	t.Setenv("CI", "")
	t.Setenv("RENDER", "")

	if defaultOpenBrowser() {
		t.Fatalf("expected browser auto-open to be disabled when PORT is present")
	}
}

func TestLoadIncludesTimeoutAndHeaderDefaults(t *testing.T) {
	t.Setenv("UNIVERSALD_APP_ENV", "staging")
	t.Setenv("UNIVERSALD_PUBLIC_ORIGIN", "https://staging.paineldief.example/")
	t.Setenv("UNIVERSALD_OPS_TOKEN", "ops-secret")
	t.Setenv("UNIVERSALD_BACKUP_RETENTION_DAYS", "21")
	t.Setenv("UNIVERSALD_MAINTENANCE_INTERVAL_SEC", "90")
	t.Setenv("UNIVERSALD_READ_TIMEOUT_SEC", "22")
	t.Setenv("UNIVERSALD_WRITE_TIMEOUT_SEC", "31")
	t.Setenv("UNIVERSALD_IDLE_TIMEOUT_SEC", "75")
	t.Setenv("UNIVERSALD_SHUTDOWN_TIMEOUT_SEC", "9")
	t.Setenv("UNIVERSALD_MAX_HEADER_BYTES", "4096")

	cfg := Load()

	if cfg.AppEnv != "staging" {
		t.Fatalf("expected app env staging, got %q", cfg.AppEnv)
	}
	if cfg.PublicOrigin != "https://staging.paineldief.example" {
		t.Fatalf("expected normalized public origin, got %q", cfg.PublicOrigin)
	}
	if cfg.OpsToken != "ops-secret" {
		t.Fatalf("expected ops token to load, got %q", cfg.OpsToken)
	}
	if cfg.BackupRetentionDays != 21 {
		t.Fatalf("expected backup retention 21, got %d", cfg.BackupRetentionDays)
	}
	if cfg.MaintenanceInterval != 90*time.Second {
		t.Fatalf("expected maintenance interval 90s, got %s", cfg.MaintenanceInterval)
	}
	if cfg.ReadTimeout != 22*time.Second {
		t.Fatalf("expected read timeout 22s, got %s", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 31*time.Second {
		t.Fatalf("expected write timeout 31s, got %s", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 75*time.Second {
		t.Fatalf("expected idle timeout 75s, got %s", cfg.IdleTimeout)
	}
	if cfg.ShutdownTimeout != 9*time.Second {
		t.Fatalf("expected shutdown timeout 9s, got %s", cfg.ShutdownTimeout)
	}
	if cfg.MaxHeaderBytes != 4096 {
		t.Fatalf("expected max header bytes 4096, got %d", cfg.MaxHeaderBytes)
	}
}

func TestResolveAppEnvDefaultsToProductionWhenHosted(t *testing.T) {
	t.Setenv("UNIVERSALD_APP_ENV", "")
	t.Setenv("PORT", "10000")

	if got := resolveAppEnv(); got != "production" {
		t.Fatalf("expected production env when PORT is present, got %q", got)
	}
}
