//go:build prod

package constants

// InstanceID is the unique identifier for the application instance.
// This should be overridden at build time using ldflags:
// -ldflags "-X 'github.com/ZenPrivacy/zen-desktop/internal/constants.InstanceID=<uuid>'"
// If not overridden, falls back to a default UUID.
var InstanceID = "a7bad8a9-cadb-4ae9-86b3-2e9e81049cb8"
