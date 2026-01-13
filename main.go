package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/getlantern/systray"
)

type Config struct {
	Port  string `xml:"Port"`
	Sites []Site `xml:"Sites>Site"`
}

type Site struct {
	Domain   string `xml:"Domain"`
	Path     string `xml:"Path"`
	AutoOpen bool   `xml:"AutoOpen"`
}

const hostMarker = "# type:fileserver-config"
const dbDir = "db"
const pidFile = "server.pid"
const logFile = "server.log"

var dbMutexes sync.Map

func main() {
	// Setup logging to file
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	log.Println("--- Server Starting with Tray ---")

	// Write PID file
	pid := os.Getpid()
	err = ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
	if err != nil {
		log.Printf("Error writing PID file: %v", err)
	} else {
		log.Printf("PID %d written to %s", pid, pidFile)
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("Go File Server")
	systray.SetTooltip("Go File Server is running")

	mQuit := systray.AddMenuItem("Quit", "Quit the server")

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	// Run server in goroutine
	go runServer()
}

func onExit() {
	log.Println("--- Server Stopping ---")
	os.Remove(pidFile)
}

func runServer() {
	// Ensure db directory exists
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		os.Mkdir(dbDir, 0755)
	}

	// Read config file
	xmlFile, err := os.Open("config.xml")
	if err != nil {
		log.Println("Error opening config.xml:", err)
		return
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var config Config
	xml.Unmarshal(byteValue, &config)

	if config.Port == "" {
		config.Port = "8080"
	}

	// Host cleanup and update
	var domains []string
	for _, site := range config.Sites {
		if site.Domain != "" {
			domains = append(domains, site.Domain)
		}
	}

	if len(domains) > 0 {
		err := updateHostsFile(domains)
		if err != nil {
			log.Printf("⚠️  Warning: Could not update hosts file: %v\n", err)
			log.Println("   (Run as Administrator to enable custom domains)")
		} else {
			log.Printf("✅ Configured %d domains in hosts file.\n", len(domains))
		}
	} else {
		_ = cleanHostsFile()
	}

	// Setup Handler
	handler := &VirtualHostHandler{Sites: config.Sites}

	log.Printf("Starting server on port %s...\n", config.Port)
	for _, site := range config.Sites {
		url := fmt.Sprintf("http://%s:%s/", site.Domain, config.Port)
		log.Printf("Serving: %s -> %s\n", url, site.Path)

		if site.AutoOpen {
			go func(u string) {
				time.Sleep(500 * time.Millisecond)
				openBrowser(u)
			}(url)
		}
	}

	err = http.ListenAndServe(":"+config.Port, handler)
	if err != nil {
		log.Fatal(err)
	}
}

// ... helper functions omitted (they remain same but need to be included in full file) ...
// To ensure full file integrity, I will include the rest of the functions below.

func openBrowser(url string) {
	var err error
	err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	if err != nil {
		log.Printf("Failed to open browser for %s: %v\n", url, err)
	}
}

type VirtualHostHandler struct {
	Sites []Site
}

func (h *VirtualHostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// API Endpoint for DB
	if strings.HasPrefix(r.URL.Path, "/api/db/") {
		handleDB(w, r)
		return
	}

	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		if strings.Contains(r.Host, ":") {
			host = r.Host
		} else {
			host = r.Host
		}
	}

	var matchedSite *Site
	for _, site := range h.Sites {
		if strings.EqualFold(site.Domain, host) {
			s := site
			matchedSite = &s
			break
		}
	}

	if matchedSite == nil {
		http.Error(w, "404 - Domain not configured", http.StatusNotFound)
		return
	}

	servePath(w, r, matchedSite.Path)
}

func handleDB(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 || parts[3] == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}
	dbName := parts[3]

	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(dbName) {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(dbDir, dbName+".json")

	mu, _ := dbMutexes.LoadOrStore(dbName, &sync.Mutex{})
	mutex := mu.(*sync.Mutex)

	mutex.Lock()
	defer mutex.Unlock()

	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			w.Write([]byte("{}"))
			return
		}
		w.Write(content)
		return
	}

	if r.Method == "POST" || r.Method == "PUT" {
		log.Printf("📝 Request to update: %s\n", dbName)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("❌ Error reading body: %v\n", err)
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		err = ioutil.WriteFile(filePath, body, 0644)
		if err != nil {
			log.Printf("❌ Error writing %s: %v\n", filePath, err)
			http.Error(w, "Failed to write db", http.StatusInternalServerError)
			return
		}

		log.Printf("✅ %s updated successfully.\n", dbName)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func servePath(w http.ResponseWriter, r *http.Request, path string) {
	fi, err := os.Stat(path)
	if err != nil {
		http.Error(w, "500 - Base path not found: "+path, http.StatusInternalServerError)
		return
	}

	if fi.IsDir() {
		fs := http.FileServer(http.Dir(path))
		fs.ServeHTTP(w, r)
		return
	} else {
		http.ServeFile(w, r, path)
	}
}

func cleanHostsFile() error {
	return updateHostsFile(nil)
}

func updateHostsFile(domains []string) error {
	hostsPath := os.Getenv("SystemRoot") + "\\System32\\drivers\\etc\\hosts"
	content, err := ioutil.ReadFile(hostsPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	for _, line := range lines {
		if strings.Contains(line, hostMarker) {
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		newLines = append(newLines, line)
	}

	if len(domains) > 0 {
		newLines = append(newLines, "")
		for _, domain := range domains {
			entry := fmt.Sprintf("127.0.0.1 %s %s", domain, hostMarker)
			newLines = append(newLines, entry)
		}
		newLines = append(newLines, "")
	}

	output := strings.Join(newLines, "\n")
	return ioutil.WriteFile(hostsPath, []byte(output), 0644)
}
