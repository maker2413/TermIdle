package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/maker2413/term-idle/internal/ssh"
)

func main() {
	// Define command line flags
	var (
		port        = flag.Int("port", 2222, "SSH server port")
		hostKeyFile = flag.String("host-key", "./ssh_host_key", "SSH host key file path")
		maxSessions = flag.Int("max-sessions", 100, "Maximum concurrent sessions")
		debug       = flag.Bool("debug", false, "Enable debug logging")
	)
	flag.Parse()

	// Set up logging
	logger := log.Default()
	if *debug {
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}

	// Create SSH server configuration
	config := &ssh.Config{
		Port:        *port,
		HostKeyFile: *hostKeyFile,
		MaxSessions: *maxSessions,
	}

	logger.Infof("Starting Term Idle SSH Server")
	logger.Infof("Port: %d", config.Port)
	logger.Infof("Host Key File: %s", config.HostKeyFile)
	logger.Infof("Max Sessions: %d", config.MaxSessions)

	// Set up graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start SSH server in a goroutine
	go func() {
		if err := ssh.StartSSHServer(config); err != nil {
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
