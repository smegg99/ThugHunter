// core/scanner/scanner.go
package scanner

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/png"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"smuggr.xyz/thughunter/common/models"
	"smuggr.xyz/thughunter/core/datastore"
)

type Result struct {
	IP       string
	Port     int
	Filename string
	Hostname string
	Labels   []string
	Location string
}

func StartControlServer() {
	http.HandleFunc("/open-vnc", func(w http.ResponseWriter, r *http.Request) {
		ip := r.URL.Query().Get("ip")
		port := r.URL.Query().Get("port")
		if ip == "" || port == "" {
			http.Error(w, "Missing ip or port", http.StatusBadRequest)
			return
		}

		template := os.Getenv("LAUNCH_VNC_COMMAND")
		if template == "" {
			http.Error(w, "LAUNCH_VNC_COMMAND not set", http.StatusInternalServerError)
			return
		}

		if strings.Count(template, "%s") < 2 {
			http.Error(w, "LAUNCH_VNC_COMMAND must contain two %s placeholders (for IP and PORT)", http.StatusInternalServerError)
			return
		}

		cmdStr := fmt.Sprintf(template, ip, port)
		cmd := exec.Command("sh", "-c", cmdStr)

		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil

		if err := cmd.Start(); err != nil {
			fmt.Printf("Failed to start VNC command: %v", err)
			http.Error(w, "Failed to start VNC client", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	addr := os.Getenv("CONTROL_SERVER_ADDR")
	if addr == "" {
		addr = "127.0.0.1:7373"
	}
	fmt.Printf("Control server listening at http://%s\n", addr)
	go http.ListenAndServe(addr, nil)
}

func RunScan(reader *bufio.Reader) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	scansDir := os.Getenv("SCANS_PATH")
	if scansDir == "" {
		scansDir = filepath.Join("scans", timestamp)
	} else {
		scansDir = filepath.Join(scansDir, timestamp)
	}

	snapshotDir := filepath.Join(scansDir, "snapshots")
	discardedDir := filepath.Join(snapshotDir, "discarded")
	os.MkdirAll(discardedDir, 0755)

	genHTML := askGenerateHTML(reader)

	var hosts []models.Host
	datastore.DB.Find(&hosts)

	working, failed, discardedCount := performParallelSnapshots(snapshotDir, discardedDir, hosts)
	writeReport(scansDir, working, failed, discardedCount)

	if genHTML {
		writeHTMLSummary(scansDir, working, failed, discardedCount)
	}
}

func askGenerateHTML(r *bufio.Reader) bool {
	fmt.Print("Generate HTML summary? (y/N): ")
	resp, _ := r.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(resp)) == "y"
}

func isSingleColorImage(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("[!] Failed to stat image: %v\n", err)
		return false
	}
	if info.Size() < 1024 {
		fmt.Printf("[!] Image too small or empty: %s\n", path)
		return false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("[!] Failed to read image: %v\n", err)
		return false
	}

	fmt.Printf("[?] Header of %s: % x\n", path, data[:12])

	contentType := http.DetectContentType(data[:512])
	if !strings.HasPrefix(contentType, "image/png") {
		fmt.Printf("[!] Invalid image type (%s): %s\n", contentType, path)
		return false
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		fmt.Printf("[!] Failed to decode image: %v\n", err)
		return false
	}

	bounds := img.Bounds()
	first := img.At(bounds.Min.X, bounds.Min.Y)
	r0, g0, b0, _ := first.RGBA()
	const tolerance = 0x0100

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if absDiff(r0, r) > tolerance || absDiff(g0, g) > tolerance || absDiff(b0, b) > tolerance {
				return false
			}
		}
	}
	return true
}

func absDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

func openFile(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}

func performParallelSnapshots(snapshotDir, discardedDir string, hosts []models.Host) ([]Result, []string, int) {
	var rawResults []Result
	var discarded int
	var failed []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	maxConcurrent := getConcurrencyLimit()
	sem := make(chan struct{}, maxConcurrent)

	for _, host := range hosts {
		port, ok := host.Services["VNC"]
		if !ok {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(h models.Host, p int) {
			defer wg.Done()
			defer func() { <-sem }()

			target := fmt.Sprintf("%s::%d", h.IP, p)
			filename := fmt.Sprintf("%s:%d.png", h.IP, p)
			output := filepath.Join(snapshotDir, filename)

			timeout := 6 * time.Second
			if toStr := os.Getenv("TIMEOUT_DEFAULT"); toStr != "" {
				if toVal, err := strconv.Atoi(toStr); err == nil && toVal > 0 {
					timeout = time.Duration(toVal) * time.Second
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			err := exec.CommandContext(ctx, "vncsnapshot", "-quiet", "-ignoreblank", target, output).Run()

			mu.Lock()
			defer mu.Unlock()

			if ctx.Err() == context.DeadlineExceeded {
				fmt.Printf("[!] %s - Timeout\n", target)
				failed = append(failed, target)
			} else if err != nil {
				fmt.Printf("[-] %s - Error: %v\n", target, err)
				failed = append(failed, target)
			} else {
				if isSingleColorImage(output) {
					fmt.Printf("[-] %s:%d - Discarded single-color image\n", h.IP, p)
					discardPath := filepath.Join(discardedDir, filename)
					os.Rename(output, discardPath)
					failed = append(failed, fmt.Sprintf("%s:%d", h.IP, p))
					discarded++
					return
				}
				fmt.Printf("[+] %s - Snapshot saved\n", target)
				rawResults = append(rawResults, Result{
					IP:       h.IP,
					Port:     p,
					Filename: filepath.Join("snapshots", filename),
					Hostname: h.Hostname,
					Labels:   h.Labels,
					Location: h.Location,
				})
			}
		}(host, port)
	}

	wg.Wait()
	return rawResults, failed, discarded
}

func getConcurrencyLimit() int {
	if maxStr := os.Getenv("MAX_CONCURRENT_VNC"); maxStr != "" {
		if val, err := strconv.Atoi(maxStr); err == nil && val > 0 {
			return val
		}
	}
	cpu := runtime.NumCPU()
	limit := cpu * 4
	if limit < 16 {
		return 16
	} else if limit > 256 {
		return 256
	}
	return limit
}

func writeReport(dir string, working []Result, failed []string, discardedCount int) {
	now := time.Now()
	dateStr := now.Format("2006-01-02_15-04-05")
	path := filepath.Join(dir, fmt.Sprintf("thug_hunting_%s.txt", dateStr))
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating report file:", err)
		return
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("VNC Thug-Hunting Report — %s\n\n", now.Format("2006-01-02 15:04:05")))
	file.WriteString(fmt.Sprintf("Total Discarded: %d\n\n", discardedCount))
	file.WriteString("Working VNC services:\n")
	for _, w := range working {
		file.WriteString(fmt.Sprintf("%s:%d\n", w.IP, w.Port))
	}
	file.WriteString("\nFailed VNC services:\n")
	for _, f := range failed {
		file.WriteString(f + "\n")
	}
	fmt.Println("📄 Report saved")
}

func writeHTMLSummary(dir string, working []Result, failed []string, discarededCount int) {
	now := time.Now()
	dateStr := now.Format("2006-01-02_15-04-05")
	path := filepath.Join(dir, fmt.Sprintf("thug_hunting_%s.html", dateStr))
	f, err := os.Create(path)
	if err != nil {
		fmt.Println("Error writing HTML summary:", err)
		return
	}
	defer f.Close()

	logoDark, _ := os.ReadFile("./assets/logo_dark.b64")
	logoLight, _ := os.ReadFile("./assets/logo_light.b64")
	darkSVG, _ := os.ReadFile("./assets/dark_mode.svg")
	lightSVG, _ := os.ReadFile("./assets/light_mode.svg")
	fontData, _ := os.ReadFile("./assets/killig.woff2.b64")
	vncConnectSVG, _ := os.ReadFile("./assets/vnc_connect.svg")

	failedCount := len(failed)
	totalCount := len(working) + failedCount
	var allHosts []models.Host
	datastore.DB.Find(&allHosts)
	totalHosts := len(allHosts)

	f.WriteString(`<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
<meta charset="UTF-8">
<title>VNC Thug-Hunter</title>
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
@font-face {
	font-family: 'KilligGang';
	src: url(data:font/woff2;base64,` + string(fontData) + `) format('woff2');
	font-display: swap;
}
h1 {
	font-family: 'KilligGang', sans-serif;
	font-size: 5rem;
	letter-spacing: 5px;
}
[data-theme="dark"] {
	--bg: #1e1e1e;
	--fg: #ffffff;
	--card-bg: #2c2c2c;
	--border: #444;
	--topbar: #111;
	--icon-fill: #ffffff;
	--btn-bg:rgba(34, 46, 58, 0);
	--btn-fg: #fff;
	--btn-hover:rgb(153, 0, 0);
}
[data-theme="light"] {
	--bg: #ffffff;
	--fg: #000000;
	--card-bg: #f0f0f0;
	--border: #ccc;
	--topbar: #f2f2f2;
	--icon-fill: #111111;
	--btn-bg:rgba(224, 231, 239, 0);
	--btn-fg: #111;
	--btn-hover:rgb(252, 179, 179);
}
body {
	background-color: var(--bg);
	color: var(--fg);
	font-family: system-ui, sans-serif;
	margin: 0;
}
header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	background-color: var(--topbar);
	color: var(--fg);
	padding: 16px;
	flex-wrap: wrap;
}
header h1 {
	font-size: 1.8rem;
	margin: 0;
}
.logo {
	height: 40px;
	margin-right: 12px;
	display: none;
}
[data-theme="dark"] .dark-logo { display: inline; }
[data-theme="light"] .light-logo { display: inline; }
.theme-toggle {
	background: none;
	border: none;
	cursor: pointer;
	width: 40px;
	height: 40px;
	padding: 0;
	display: flex;
	align-items: center;
	justify-content: center;
}
.theme-toggle svg {
	width: 28px;
	height: 28px;
	fill: var(--icon-fill);
}
[data-theme="light"] .toggle-dark { display: none; }
[data-theme="dark"] .toggle-light { display: none; }
.stats {
	padding: 12px 16px;
	font-size: 0.95rem;
	background: var(--card-bg);
	border-bottom: 1px solid var(--border);
}
button {
	background-color: var(--btn-bg);
	color: var(--btn-fg);
	padding: 6px 12px;
	border: none;
	border-radius: 4px;
	cursor: pointer;
	font-size: 0.85rem;
	transition: background 0.2s ease-in-out, color 0.2s;
}
button:hover {
	background-color: var(--btn-hover);
	color: var(--btn-fg);
}
.grid {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
	gap: 12px;
	padding: 12px;
}
.card {
	background: var(--card-bg);
	border: 1px solid var(--border);
	border-radius: 8px;
	padding: 12px;
	display: flex;
	flex-direction: column;
}
.card h2 {
	font-size: 1rem;
	margin: 0 0 6px 0;
	word-wrap: break-word;
}
.card p {
	margin: 4px 0;
	font-size: 0.85rem;
}
.card img {
	width: 100%;
	height: auto;
	margin-top: auto;
	border: 1px solid #555;
	border-radius: 4px;
	cursor: zoom-in;
}
#overlay {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background-color: rgba(0, 0, 0, 0.85);
	display: none;
	align-items: center;
	justify-content: center;
	z-index: 9999;
}
#overlay img {
	max-width: 95%;
	max-height: 95%;
	box-shadow: 0 0 12px #000;
	border-radius: 6px;
	border: 2px solid white;
}
.vnc-icon-button {
	position: absolute;
	top: 8px;
	right: 8px;
	width: 28px;
	height: 28px;
	background-color: var(--btn-bg);
	border-radius: 6px;
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	transition: background-color 0.2s ease-in-out;
}
.vnc-icon-button svg {
	width: 20px;
	height: 20px;
	fill: var(--icon-fill);
}
.vnc-icon-button:hover {
	background-color: var(--btn-hover);
}
.card {
	position: relative;
}
</style>
</head>
<body>
<header>
	<div style="display: flex; align-items: center;">
		<img class="logo dark-logo" src="data:image/png;base64,` + string(logoDark) + `" alt="Logo Dark">
		<img class="logo light-logo" src="data:image/png;base64,` + string(logoLight) + `" alt="Logo Light">
		<h1>Da Thug-Hunting Summary</h1>
	</div>
	<button class="theme-toggle" onclick="toggleTheme()">
		<span class="toggle-dark">` + string(darkSVG) + `</span>
		<span class="toggle-light">` + string(lightSVG) + `</span>
	</button>
</header>

<div class="stats">
	<div><strong>Report Date:</strong> ` + now.Format("2006-01-02 15:04:05") + `</div>
	<div><strong>Total Hosts:</strong> ` + strconv.Itoa(totalHosts) + ` |
	<strong>Succeeded:</strong> ` + strconv.Itoa(totalCount - failedCount - discarededCount) + ` |
	<strong>Failed:</strong> ` + strconv.Itoa(failedCount) + ` |
	<strong>Discarded:</strong> ` + strconv.Itoa(discarededCount) + `</div>
	</div>
</div>

<div class="grid">`)

	for _, r := range working {
		f.WriteString(`<div class="card">`)
		f.WriteString(fmt.Sprintf("<h2>%s:%d</h2>", r.IP, r.Port))
		if r.Hostname != "" {
			f.WriteString(fmt.Sprintf("<p><strong>Hostname:</strong> %s</p>", r.Hostname))
		}
		if r.Location != "" {
			f.WriteString(fmt.Sprintf("<p><strong>Location:</strong> %s</p>", r.Location))
		}
		f.WriteString(fmt.Sprintf(`<img src="%s" alt="Snapshot of %s">`, r.Filename, r.IP))
		f.WriteString(fmt.Sprintf(`
		<div class="vnc-icon-button" onclick="launchVNC('%s', '%d')" title="Connect via VNC">
			%s
		</div>
		`, r.IP, r.Port, string(vncConnectSVG)))
		f.WriteString(`</div>`)
	}

	f.WriteString(`</div>
<div id="overlay" onclick="hideOverlay()">
	<img id="overlay-img" src="" alt="Fullscreen">
</div>

<script>
const CONTROL_SERVER_ADDR = "` + os.Getenv("CONTROL_SERVER_ADDR") + `";
function getControlServerURL() {
	let addr = CONTROL_SERVER_ADDR || "127.0.0.1:7373";
	if (!addr.startsWith("http")) addr = "http://" + addr;
	return addr;
}
function toggleTheme() {
	const html = document.documentElement;
	html.setAttribute("data-theme", html.getAttribute("data-theme") === "dark" ? "light" : "dark");
}
function showOverlay(src) {
	document.getElementById("overlay-img").src = src;
	document.getElementById("overlay").style.display = "flex";
}
function hideOverlay() {
	document.getElementById("overlay").style.display = "none";
}
function launchVNC(ip, port) {
	const url = getControlServerURL() + '/open-vnc?ip=' + ip + '&port=' + port;
	fetch(url)
		.then(r => {
			if (r.ok) {
				alert("Launching VNC viewer...");
			}
			return;
		})
		.catch(err => {
			console.warn("Fetch to control server failed (VNC may still launch):", err);
		});
}
document.addEventListener("keydown", e => {
	if (e.key === "Escape") hideOverlay();
});
document.querySelectorAll(".card img").forEach(img => {
	img.onclick = e => {
		e.stopPropagation();
		showOverlay(img.src);
	};
});
</script>
</body>
</html>`)

	err = openFile(path)
	if err != nil {
		fmt.Println("Failed to open generated file:", err)
	}

	fmt.Println("HTML summary saved")
}
