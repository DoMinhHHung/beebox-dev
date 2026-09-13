package capability

import (
	"github.com/DoMinhHHung/beebox-dev/modules/beebox-auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/domain"
)

type Key struct {
	ModuleID          string
	ModuleVersion     string
	CapabilityID      string
	CapabilityVersion string
}

func KeyFromApplied(cfg domain.AppliedConfiguration) Key {
	return Key{
		ModuleID:          cfg.ModuleID,
		ModuleVersion:     cfg.ModuleVersion,
		CapabilityID:      cfg.CapabilityID,
		CapabilityVersion: cfg.CapabilityVersion,
	}
}

func RequireSessionAuth(cfg domain.AppliedConfiguration) error {
	if cfg.AppliedVersion < 1 {
		return apperror.New(apperror.CodeNotFound, "configuration not applied")
	}
	key := KeyFromApplied(cfg)
	if key.ModuleID != beeboxauth.ModuleID {
		return apperror.New(apperror.CodeForbidden, "auth capability is not applied")
	}
	if key.ModuleVersion != beeboxauth.ModuleVersion {
		return apperror.New(apperror.CodeConflict, "unsupported auth module version")
	}
	if key.CapabilityID != beeboxauth.CapabilitySession {
		return apperror.New(apperror.CodeForbidden, "session capability is not applied")
	}
	if key.CapabilityVersion != beeboxauth.CapabilitySessionVersion {
		return apperror.New(apperror.CodeConflict, "unsupported session capability version")
	}
	return nil
}
