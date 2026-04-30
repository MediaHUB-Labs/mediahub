package main

import "embed"

// Embed only the runtime UI files into the binary.
// This excludes dev-only files (tailwindcss binary, input.css, tailwindBuild.sh)
// and keeps the binary size small (~1MB of UI assets vs 115MB+ with everything).
//
//go:embed mediahub-ui/index.html
//go:embed mediahub-ui/App.js
//go:embed mediahub-ui/view
//go:embed mediahub-ui/src
//go:embed mediahub-ui/assets
var embeddedUI embed.FS
