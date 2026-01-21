package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/maker2413/term-idle/internal/config"
	"github.com/maker2413/term-idle/internal/ssh"
)

func main() {
	var configPath = flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up logging
	logger := log.Default()
	switch cfg.Logging.Level {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}

	// Create SSH server configuration
	sshConfig := &ssh.Config{
		Port:        cfg.SSH.Port,
		HostKeyFile: cfg.SSH.HostKeyFile,
		MaxSessions: cfg.SSH.MaxSessions,
	}

	logger.Infof("Starting Term Idle SSH Server")
	logger.Infof("Port: %d", sshConfig.Port)
	logger.Infof("Host Key File: %s", sshConfig.HostKeyFile)
	logger.Infof("Max Sessions: %d", sshConfig.MaxSessions)

	// Set up graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start SSH server in a goroutine
	go func() {
		if err := ssh.StartSSHServer(sshConfig); err != nil {
			logger.Fatalf("Failed to start SSH server: %v", err)
		}
	}()

	logger.Info("SSH server started successfully")
	logger.Info("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	<-done
	logger.Info("Shutting down SSH server...")

	// TODO: Add graceful shutdown logic
	logger.Info("Server stopped")
}
