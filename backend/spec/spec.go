// Package spec embeds the OpenAPI specification for the RV Logbook API.
// It is imported by the HTTP server to serve the spec at /openapi.yaml and
// by the Scalar UI route at /docs.
package spec

import _ "embed"

// OpenAPI contains the raw bytes of openapi.yaml, embedded at compile time.
// Serving it from the binary means the spec and the running code are always in sync.
//
//go:embed openapi.yaml
var OpenAPI []byte
