// core/scanner/scanner.go
package scanner

import (
	"bufio"
	"context"
	"fmt"
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

func RunScan(reader *bufio.Reader) {
	snapshotDir := getSnapshotDir()
	genHTML := askGenerateHTML(reader)

	var hosts []models.Host
	datastore.DB.Find(&hosts)

	working, failed := performParallelSnapshots(snapshotDir, hosts)
	writeReport(snapshotDir, working, failed)

	if genHTML {
		writeHTMLSummary(snapshotDir, working)
	}
}

func getSnapshotDir() string {
	dir := os.Getenv("SNAPSHOTS_PATH")
	if dir == "" {
		dir = "./snapshots"
	}
	os.MkdirAll(dir, 0755)
	return dir
}

func askGenerateHTML(r *bufio.Reader) bool {
	fmt.Print("Generate HTML summary? (y/N): ")
	resp, _ := r.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(resp)) == "y"
}

func performParallelSnapshots(snapshotDir string, hosts []models.Host) ([]Result, []string) {
	var working []Result
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

			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()

			err := exec.CommandContext(ctx, "vncsnapshot", "-quiet", "-nojpeg", "-compresslevel", "0", target, output).Run()

			mu.Lock()
			if ctx.Err() == context.DeadlineExceeded {
				fmt.Printf("[!] %s - Timeout\n", target)
				failed = append(failed, target)
			} else if err == nil {
				fmt.Printf("[+] %s - Snapshot saved\n", target)
				working = append(working, Result{IP: h.IP, Port: p, Filename: filename, Hostname: h.Hostname, Labels: h.Labels, Location: h.Location})
			} else {
				fmt.Printf("[-] %s - Error: %v\n", target, err)
				failed = append(failed, target)
			}
			mu.Unlock()
		}(host, port)
	}

	wg.Wait()
	return working, failed
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

func writeReport(dir string, working []Result, failed []string) {
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

func writeHTMLSummary(dir string, results []Result) {
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

	total := len(results)
	var allHosts []models.Host
	datastore.DB.Find(&allHosts)
	totalHosts := len(allHosts)
	failed := totalHosts - total

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
}
[data-theme="light"] {
	--bg: #ffffff;
	--fg: #000000;
	--card-bg: #f0f0f0;
	--border: #ccc;
	--topbar: #f2f2f2;
	--icon-fill: #111111;
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
	<strong>VNC Scanned:</strong> ` + strconv.Itoa(total) + ` |
	<strong>Failed:</strong> ` + strconv.Itoa(failed) + `</div>
</div>

<div class="grid">`)

	for _, r := range results {
		f.WriteString(`<div class="card">`)
		f.WriteString(fmt.Sprintf("<h2>%s:%d</h2>", r.IP, r.Port))
		if r.Hostname != "" {
			f.WriteString(fmt.Sprintf("<p><strong>Hostname:</strong> %s</p>", r.Hostname))
		}
		if r.Location != "" {
			f.WriteString(fmt.Sprintf("<p><strong>Location:</strong> %s</p>", r.Location))
		}
		f.WriteString(fmt.Sprintf(`<img src="%s" alt="Snapshot of %s">`, r.Filename, r.IP))
		f.WriteString(`</div>`)
	}

	f.WriteString(`</div>
<div id="overlay" onclick="hideOverlay()">
	<img id="overlay-img" src="" alt="Fullscreen">
</div>

<script>
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

	fmt.Println("HTML summary saved")
}
