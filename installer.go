package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	installDir  = "/usr/local/bin/mediahub"
	serviceName = "mediahub"
	binaryName  = "mediahub"
)

// handleInstall copies the binary to installDir, creates a .env with a
// generated JWT secret, writes a systemd unit, and enables + starts it.
func handleInstall() {
	if os.Getuid() != 0 {
		fmt.Println("❌ Installation requires root privileges.")
		fmt.Println("   Run: sudo ./mediahub install")
		os.Exit(1)
	}

	actualUser := os.Getenv("SUDO_USER")
	if actualUser == "" {
		actualUser = "root"
	}

	fmt.Println("🚀 Installing MediaHUB...")

	// ── 1. Create install directory ──────────────────────────
	if err := os.MkdirAll(installDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// ── 2. Copy binary ──────────────────────────────────────
	selfPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to resolve own path: %v", err)
	}
	selfPath, _ = filepath.EvalSymlinks(selfPath)

	destPath := filepath.Join(installDir, binaryName)
	if selfAbs, _ := filepath.Abs(selfPath); selfAbs != destPath {
		if err := copyFile(selfPath, destPath); err != nil {
			log.Fatalf("Failed to copy binary: %v", err)
		}
		os.Chmod(destPath, 0755)
		fmt.Printf("   ✔ Binary → %s\n", destPath)
	} else {
		fmt.Println("   ✔ Binary already in place")
	}

	// ── 3. Create .env (only if missing) ─────────────────────
	envPath := filepath.Join(installDir, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		secret := generateSecret(32)
		env := fmt.Sprintf("APP_PORT=9123\nDB_PATH=%s/mediahub.db\nUPLOAD_PATH=%s/uploads\nJWT_SECRET=%s\nGIN_MODE=release\n",
			installDir, installDir, secret)
		if err := os.WriteFile(envPath, []byte(env), 0600); err != nil {
			log.Fatalf("Failed to create .env: %v", err)
		}
		fmt.Printf("   ✔ Config  → %s\n", envPath)
	} else {
		fmt.Println("   ✔ Config already exists, skipping")
	}

	// ── 4. Set ownership ─────────────────────────────────────
	runCmd("chown", "-R", actualUser+":"+actualUser, installDir)

	// ── 5. Create systemd service ────────────────────────────
	unit := fmt.Sprintf(`[Unit]
Description=MediaHUB - Local Media Server
After=network.target

[Service]
Type=simple
User=%s
WorkingDirectory=%s
ExecStart=%s
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`, actualUser, installDir, destPath)

	svcPath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	if err := os.WriteFile(svcPath, []byte(unit), 0644); err != nil {
		log.Fatalf("Failed to create service file: %v", err)
	}
	fmt.Printf("   ✔ Service → %s\n", svcPath)

	// ── 6. Enable & start ────────────────────────────────────
	runCmd("systemctl", "daemon-reload")
	runCmd("systemctl", "enable", serviceName)
	runCmd("systemctl", "restart", serviceName)

	fmt.Println()
	fmt.Println("✅ MediaHUB installed and running!")
	fmt.Printf("   🌐 http://localhost:9123\n")
	fmt.Printf("   📂 Data:   %s\n", installDir)
	fmt.Printf("   🔧 Config: %s\n", envPath)
	fmt.Println()
	fmt.Printf("   systemctl status %s     # check status\n", serviceName)
	fmt.Printf("   systemctl restart %s    # restart\n", serviceName)
	fmt.Printf("   sudo %s uninstall       # remove\n", destPath)
}

// handleUninstall stops the service, removes the unit file and binary,
// but preserves the data directory (DB + uploads).
func handleUninstall() {
	if os.Getuid() != 0 {
		fmt.Println("❌ Uninstallation requires root privileges.")
		fmt.Println("   Run: sudo ./mediahub uninstall")
		os.Exit(1)
	}

	fmt.Println("🗑️  Uninstalling MediaHUB...")

	runCmd("systemctl", "stop", serviceName)
	runCmd("systemctl", "disable", serviceName)

	svcPath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	os.Remove(svcPath)
	runCmd("systemctl", "daemon-reload")
	fmt.Println("   ✔ Service removed")

	os.Remove(filepath.Join(installDir, binaryName))
	fmt.Println("   ✔ Binary removed")

	fmt.Println()
	fmt.Println("✅ MediaHUB uninstalled!")
	fmt.Printf("   📂 Data preserved at: %s\n", installDir)
	fmt.Printf("   To remove all data:  sudo rm -rf %s\n", installDir)
}

// ── helpers ──────────────────────────────────────────────────

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func generateSecret(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func runCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
