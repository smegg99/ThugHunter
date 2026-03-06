web_dir := "frontend"
app_name := "thughunter"
bin_dir := "bin"
config_file := "./build/config.yml"
vite_port := env("WAILS_VITE_PORT", "9245")

# List available commands
default:
    @just --list

# Run the full Wails app in development mode (hot-reload)
dev:
    wails3 dev -config {{config_file}} -port {{vite_port}}

# Start only the Nuxt frontend dev server
dev-frontend:
    cd {{web_dir}} && pnpm dev

# Production build (current platform)
build:
    wails3 build

# Development build (fast, unoptimised)
build-dev:
    wails3 build -devbuild

# Debug build (includes symbols)
build-debug:
    wails3 build -debug

# Clean build (remove cache first)
build-clean:
    wails3 build -clean

# Build for a specific platform (e.g. just build-platform linux/amd64)
build-platform platform:
    wails3 build -platform {{platform}}

# Package the app for the current platform
package:
    wails3 package

# Package for a specific OS (e.g. just package-os linux)
package-os os:
    wails3 package GOOS={{os}}

# Create Linux AppImage
create-appimage:
    wails3 task linux:create:appimage

# Run the built application
run:
    ./{{bin_dir}}/{{app_name}}

# Generate frontend bindings from Go services
generate-bindings:
    wails3 generate bindings -ts -clean=true

# Generate frontend bindings as TypeScript interfaces
generate-bindings-interfaces:
    wails3 generate bindings -ts -i

# Generate app icons from build/appicon.png
generate-icons:
    wails3 generate icons -input build/appicon.png -windowsfilename build/windows/icon.ico -iconcomposerinput build/appicon.icon

# Generate Go code from CUE schemas
generate-cue:
    go generate ./...

# Update build assets from config
update-build-assets:
    wails3 update build-assets -name "{{app_name}}" -binaryname "{{app_name}}" -config config.yml -dir build