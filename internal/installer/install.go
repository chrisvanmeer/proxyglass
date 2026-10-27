package installer

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"

	"github.com/chrisvanmeer/proxyglass/internal/config"
	"github.com/chrisvanmeer/proxyglass/internal/logger"
)

const (
	serviceName = "proxyglass"
	serviceUser = "proxyglass"
	configDir   = "/etc/proxyglass"
	dataDir     = "/var/lib/proxyglass"
	logDir      = "/var/log/proxyglass"
	configPath  = "/etc/proxyglass/config.yml"
	systemdPath = "/etc/systemd/system/proxyglass.service"
)

func RunInstall() error {
	logger.Info("Installing Proxyglass service, user, and configuration...")

	if _, err := user.Lookup(serviceUser); err != nil {
		if _, ok := err.(user.UnknownUserError); ok {
			if _, groupErr := user.LookupGroup(serviceUser); groupErr != nil {
				if err := runCmd("groupadd", "-r", serviceUser); err != nil {
					return fmt.Errorf("failed to create group: %w", err)
				}
			}
			logger.Info("Creating user %s", serviceUser)
			if err := runCmd("useradd", "-r", "-g", serviceUser, "-s", "/sbin/nologin", "-d", dataDir, "-m", serviceUser); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			return err
		}
	} else {
		logger.Info("User %s already exists", serviceUser)
	}

	for _, dir := range []string{configDir, dataDir, logDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	if u, err := user.Lookup(serviceUser); err == nil {
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)
		_ = os.Chown(configDir, 0, gid)
		_ = os.Chown(dataDir, uid, gid)
		_ = os.Chown(logDir, uid, gid)
		_ = os.Chmod(configDir, 0750)
		_ = os.Chmod(dataDir, 0750)
		_ = os.Chmod(logDir, 0750)
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	targetBin := "/usr/bin/proxyglass"
	if err := copyFile(exe, targetBin, 0755); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Info("Creating default configuration at %s", configPath)
		if err := os.WriteFile(configPath, []byte(config.GetDefaultConfigYML()), 0640); err != nil {
			return err
		}
		if u, err := user.Lookup(serviceUser); err == nil {
			gid, _ := strconv.Atoi(u.Gid)
			_ = os.Chown(configPath, 0, gid)
		}
		_ = os.Chmod(configPath, 0640)
	} else {
		logger.Info("Configuration already exists at %s", configPath)
	}

	serviceContent := `[Unit]
Description=Proxyglass HAProxy Stats Dashboard
After=network.target

[Service]
Type=simple
User=proxyglass
Group=proxyglass
EnvironmentFile=-/etc/default/proxyglass
ExecStart=/usr/bin/proxyglass --config ${PROXYGLASS_CONFIG}
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
`
	if err := os.WriteFile(systemdPath, []byte(serviceContent), 0644); err != nil {
		return err
	}

	defaultPath := "/etc/default/proxyglass"
	if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
		def := "# Proxyglass service environment\nPROXYGLASS_CONFIG=/etc/proxyglass/config.yml\n# HTTP_PROXY=http://proxy.example.net:3128\n# HTTPS_PROXY=http://proxy.example.net:3128\n# NO_PROXY=localhost,127.0.0.1,.internal\n"
		if err := os.WriteFile(defaultPath, []byte(def), 0644); err != nil {
			return err
		}
	}

	_ = runCmd("systemctl", "daemon-reload")

	logger.Info("Installation complete!")
	logger.Info("  1. Edit /etc/proxyglass/config.yml to configure HAProxy targets")
	logger.Info("  2. Generate an admin token: proxyglass --generate-token --config /etc/proxyglass/config.yml")
	logger.Info("  3. Start the service: systemctl start proxyglass")
	logger.Info("  4. Enable on boot: systemctl enable proxyglass")

	return nil
}

func RunUninstall() error {
	logger.Info("Uninstalling Proxyglass...")
	_ = runCmd("systemctl", "stop", serviceName)
	_ = runCmd("systemctl", "disable", serviceName)
	_ = os.Remove(systemdPath)
	_ = runCmd("systemctl", "daemon-reload")
	_ = os.Remove("/usr/bin/proxyglass")
	_ = os.Remove("/etc/default/proxyglass")
	logger.Info("Proxyglass uninstalled. Configuration and data in /etc/proxyglass and /var/lib/proxyglass left in place.")
	return nil
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string, mode os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, mode)
}
