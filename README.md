<p align="center">
  <img src="assets/thethughunter.svg" alt="ThugHunter" width="400">
</p>

ThugHunter is a desktop application that aggregates various publicly accessible hosts from Censys. Currently focused on VNC, with support for additional protocols planned.

Built with Go, Wails 3, and Nuxt (Vue 3).

## What It Does

ThugHunter scrapes hosts data from Censys using automated stealth browser sessions  ([Unrevealed](https://github.com/smegg99/Unrevealed) + [Human](https://github.com/smegg99/Human)) backed by a configurable pool of agents. Each agent manages its own Censys account, and the app can handle account registration, login, and credit tracking automatically. Email verification for new accounts is handled through an IMAP catch-all mailbox.

Collected hosts are pinged concurrently to check whether they are still alive. For VNC endpoints, the app captures screenshots using a native RFB implementation with an external tool fallback. All host records, including geolocation, OS, hardware info, latency, labels, and associated services, are stored in a local SQLite database.

The app runs as a desktop application with system tray integration. English and Polish locales are supported.

### Protocol Support

- [x] VNC
- [ ] RDP
- [ ] SPICE
- [ ] Modbus
- [ ] Pavel Khlebovich camera app

## Prerequisites

Linux only for now. Windows and macOS support is planned but not yet available.

### Build Dependencies

- Go 1.25+ (with CGO enabled)
- Node.js + pnpm
- C compiler (gcc or clang)
- pkg-config
- GTK 3 dev headers (`libgtk-3-dev` on Debian/Ubuntu, `gtk3-devel` on Fedora)
- WebKit2GTK 4.1 dev headers (`libwebkit2gtk-4.1-dev` on Debian/Ubuntu, `webkit2gtk4.1-devel` on Fedora)
- Wails 3 CLI (`go install github.com/wailsapp/wails/v3/cmd/wails3@latest`)

### Runtime Dependencies

- Chromium or Chrome (for the scraper, I recommend ungoogled chromium as it is quite lightweight)
- Xvfb (virtual display for running the browser without a visible window, headless mode is not used because it triggers bot detection, therefore windows and macOS users are at a disadvantage here)
- vncdo (as a fallback, captures VNC screenshots)
- xdg-open (opens HTTP, HTTPS, RDP, and SSH links in default applications)
- remmina (default VNC client, works for RDP/SPICE as well)

### Optional

- Task or Just (command runners, either works, both have recipes defined, the justfile is mine, idk if I will leave decide to keep it as it adds unnecessary complexity)

## Setup

### 1. Clone the repository

```sh
git clone https://github.com/smegg99/ThugHunter
cd ThugHunter
```

### 2. Install frontend dependencies

```sh
cd frontend && pnpm install && cd ..
```

### 3. Start the app

```sh
just dev
```

On first launch, the app will generate a default `config.json` and a DB directory (if they're not present). You can edit the config file via a text editor or just use the settings page in the UI. See the [Configuration](#configuration) section below for details on available settings.

## Development

```sh
# Full app with hot-reload (Wails + Nuxt)
just dev
```

## Building

Builds are created in the `bin/` directory. To run the app, execute the generated binary directly or use the provided `just run` command.

```sh
# Production build (current platform)
just build

# Dev build (fast, no optimization)
just build-dev
```

## Packaging

```sh
# Package for current OS
just package

# Package for a specific OS
just package-os linux

# Linux AppImage
just create-appimage
```

## Configuration

Config is loaded from `config.json` next to the binary, or from `~/.config/thughunter/config.json`.

The schema is defined in `schema/config.cue`.

Config values support dynamic references using `@{source:key}` syntax (look at the `common/config` package). Adding `?` after the source name (e.g. `@{env?:VAR}`) makes resolution optional, returning an empty string instead of an error if the lookup fails.

- `@{env:VAR}` reads an OS environment variable
- `@{keyring:key}` reads from the OS keyring under the `thughunter` service

### Scanner Options

| Option | Default | Description |
|---|---|---|
| `scanner.ping_mode` | `soft` | Controls whether services are probed when ping fails. `soft` probes all services regardless of ping result (slower, but catches services on hosts that block ICMP). `strict` skips service probes entirely if the host doesn't respond to ping (faster, may miss live services). |
| `scanner.icmp_ping` | `true` | When `true`, sends an ICMP echo request first. If ICMP fails, falls back to TCP connect on ports 80, 443, 7. When `false`, skips ICMP and goes straight to TCP fallback. Set to `false` if you don't have ICMP permissions. |
| `scanner.reject_blank_screenshots` | `true` | Rejects screenshots that appear to be a single solid color (per-channel pixel variance below 12.0, sampled across up to 10k pixels). Useful for filtering out blank screens. Set to `false` to keep all captures. |
| `scanner.workers.max_workers` | `2000` | Number of concurrent goroutines for pinging hosts and probing services. Automatically capped at the number of hosts being scanned. High values can overwhelm your network. |
| `scanner.workers.screenshot_max_workers` | `32` | Semaphore size controlling how many VNC screenshot processes (both native and external) can run at the same time. Shared between native RFB capture and the external tool fallback. |
| `scanner.workers.ping_timeout_seconds` | `3` | Timeout for each ping attempt (ICMP echo or individual TCP connect on fallback ports). |
| `scanner.workers.connect_timeout_seconds` | `10` | TCP dial timeout when connecting to a VNC service for probing. |
| `scanner.workers.banner_timeout_seconds` | `10` | Read deadline for the RFB protocol handshake after a VNC connection is established. If the server doesn't send the version banner within this time, the probe fails. |
| `scanner.workers.screenshot_timeout_seconds` | `15` | Total time budget for capturing a screenshot. Covers both the native RFB capture attempt and, if needed, the external tool fallback. |
| `scanner.workers.screenshot_delay_seconds` | `1` | Delay in seconds before the external screenshot tool captures. Passed to the command template as `{{.DELAY}}`. Gives the remote display time to finish rendering. |
| `scanner.workers.screenshot_pause_seconds` | `5` | Pause argument passed to the external tool (e.g. `vncdo pause`). Tells the VNC client to wait before taking the snapshot. Used when native capture returns a blank image. |

If the native way of capturing screenshot fails or produces a blank image, the external command template (`scanner.templates.screenshot_command`) is used as fallback.

### Scraper Options

| Option | Default | Description |
|---|---|---|
| `scraper.agents.max_agents` | `10` | Maximum number of simultaneous Chrome browser instances. Each agent is assigned one Censys account, opens a browser, and executes search queries. More agents means more parallel searches but higher memory usage. |
| `scraper.agents.max_register_retries` | `3` | How many times to retry a failed account registration before giving up. |
| `scraper.browser_binary_path` | `/usr/bin/google-chrome-stable` | Absolute path to the Chromium or Chrome binary used by the scraper. |
| `scraper.virtual_display` | `true` | When `true`, each agent spawns its own Xvfb virtual display. This avoids showing browser windows on screen and prevents X11 connection exhaustion when running many agents in parallel. When `false`, uses the system display. |
| `scraper.minimal_browser` | `false` | When `true`, disables extra browser features and plugins for a smaller memory footprint. |
| `scraper.custom_query_strings` | `[]` | Additional Censys search queries beyond the built-in ones. Supports `{{ALL_COUNTRIES}}` (expands to one query per country, ~97 queries) and `{{CONTINENT:Name}}` (expands per country in that continent). Available continents: Europe, Asia, North America, South America, Africa, Oceania, Middle East.

### Account Credential Templates

When the scraper registers new Censys accounts, it fills form fields using configurable templates under `scraper.agents.templates`. Each template is resolved with the following variables, generated per registration using gofakeit:

| Variable | Description |
|---|---|
| `{{.ACCOUNT_ID}}` | Agent name + timestamp in milliseconds |
| `{{.RANDOM_NONSENSE}}` | 8-character random string (mixed case, digits, special chars) |
| `{{.FIRST_NAME}}` / `{{.LAST_NAME}}` | Random first/last name |
| `{{.LC_FIRST_NAME}}` / `{{.LC_LAST_NAME}}` | Lowercase variants |
| `{{.FULL_NAME}}` / `{{.LC_FULL_NAME}}` | Combined full name and lowercase variant |
| `{{.USERNAME}}` | First letter + last name + random number (e.g. `jsmith47`) |
| `{{.COMPANY}}` | Random company name |
| `{{.CITY}}` / `{{.COUNTRY}}` | Random city and country |
| `{{.JOB_TITLE}}` / `{{.BUZZWORD}}` / `{{.DOMAIN}}` | Random job title, buzzword, domain |
| `{{.DIGITS_1}}` through `{{.DIGITS_6}}` | Random number with 1 to 6 digits |

### Scanner Command Templates

Commands for connecting to services and capturing screenshots are configurable under `scanner.templates`. These are resolved with `{{.IP}}` and `{{.PORT}}` at runtime.

The screenshot command template has additional variables: `{{.TIMEOUT}}`, `{{.DELAY}}`, `{{.PAUSE}}` (from scanner worker config), and `{{.OUTPUT}}` (temp file path for the captured image).

## Platforms

- Linux (native binary, AppImage, rpm, deb, AUR) (fully supported, tested on EndeavourOS with KDE Plasma, Wayland)
- Windows (planned, build scaffolding exists but untested)
- macOS (planned, pending access to hardware)

## Disclaimer

This software is provided as-is, without warranty of any kind. The author is not responsible for any damages, legal consequences, or misuse arising from the use of this application. Use it at your own risk and ensure compliance with all applicable laws and regulations in your jurisdiction.

## License

See [LICENSE](LICENSE).
