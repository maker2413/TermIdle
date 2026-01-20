# AGENTS.md

This file contains guidelines and commands for agentic coding agents working on the Term Idle project.

## Project Overview

Term Idle is a terminal-based idle game. The project is currently in early development with minimal infrastructure.

## Build/Lint/Test Commands

### Primary Commands
- **Build**: `make build` - Compiles the binary to `./term-idle`
- **Run**: `make run` - Runs the application with config from `cmd/term-idle/main.go`
- **Lint**: `make lint` - Runs golangci-lint
- **Test**: `make test` - Runs tests with coverage for all packages

### Manual Commands
- **Run single test**: `go test -v ./path/to/package -run TestSpecificFunction`
- **Run package tests**: `go test -v ./internal/player`
- **Build without makefile**: `go build -o term-idle cmd/term-idle/main.go`
- **Run without makefile**: `go run cmd/term-idle/main.go`

### Dependencies
- Install dependencies: `go mod tidy`
- Download dependencies: `go mod download`
- Verify dependencies: `go mod verify`

## Code Style Guidelines

### Imports
- Group imports in three sections: standard library, third-party, internal packages
- Use blank lines between groups
- Example:
  ```go
  import (
      "log"
      "time"
      
      "github.com/charmbracelet/bubbletea"
      "github.com/gopxl/beep"
      
      "github.com/maker2413/go-radio-player/internal/config"
      "github.com/maker2413/go-radio-player/internal/player"
  )
  ```

### Naming Conventions
- **Packages**: lowercase, single words when possible (`game`, `config`, `player`)
- **Constants**: `UpperCamelCase` for exported constants, `lowerSnake` for internal
- **Variables**: `camelCase` for all variables
- **Functions**: `CamelCase` with export visibility
- **Interfaces**: Usually end with `-er` suffix when possible (`io.Reader`, `beep.Streamer`)
- **Struct fields**: `CamelCase` when exported, `camelCase` when unexported

### Error Handling
- Always handle errors immediately, never ignore them
- Use `log.Fatal(err)` for unrecoverable errors in main
- For package functions, return errors and let callers handle them
- Use `fmt.Errorf` for wrapping errors with context
- Example pattern:
  ```go
  if err != nil {
      return nil, fmt.Errorf("failed to create game: %w", err)
  }
  ```

### Concurrency
- Use mutexes for protecting shared state (`sync.Mutex`)
- Channel communication should be non-blocking in updates with select/default
- Use `speaker.Lock()`/`speaker.Unlock()` for audio stream operations

### Structs and Interfaces
- Define receiver variables with first letter of type: `(ap *audioPlayer)`
- Use constructor functions: `NewAudioPlayer(...)` returning pointer and error
- Keep unexported fields to maintain encapsulation
- Use dependency injection for testability

### Constants and Magic Numbers
- Define constants at package level with descriptive names
- Group related constants
- Avoid magic numbers in code - use named constants

### TUI (Bubbletea) Patterns
- Implement `Init()`, `Update(msg tea.Msg)`, and `View()` methods
- Use `tea.Cmd` for async operations
- Handle `tea.WindowSizeMsg` for responsive layouts
- Use `lipgloss` for styling with predefined styles
- Support standard key bindings: `ctrl+c`, `esc` for quit

### File Organization
- `cmd/` - Main applications entry points
- `internal/` - Private application code
  - `config/` - Configuration management
  - `player/` - Player logic
  - `game/` - Game logic
- Tests should be in the same package as the code they test (`*_test.go`)

### Configuration
- Use `koanf` for configuration management
- Support both YAML files and environment variables
- Provide sensible defaults
- Validate configuration on startup

### Comments and Documentation
- Exported functions should have godoc comments
- Keep comments concise and focused on "why" not "what"
- Use TODO/FIXME comments sparingly

## Development Setup

### Prerequisites
- Go 1.25.5+ (as specified in go.mod)
- libasound2-dev for audio support on Linux

### Environment
- Create `config.yaml` based on `config.yaml` example
- Use `.env.example` as template for environment variables
- Debug logging writes to `debug.log` when enabled

### Testing
- Currently no test files exist - tests should be added for all packages
- Use testify for assertions
- Aim for good test coverage on business logic

## Key Dependencies

- **github.com/charmbracelet/bubbletea**: TUI framework
- **github.com/gopxl/beep**: Audio playback
- **github.com/knadh/koanf**: Configuration management
- **github.com/charmbracelet/lipgloss**: TUI styling

## Common Patterns

### Resource Management
- Use `defer` for cleanup operations
- Always close `io.Closer` interfaces
- Handle cleanup in defer functions with error checking

### Channel Patterns
- Use buffered channels to prevent blocking
- Use select statements with default cases for non-blocking operations
- Close channels when appropriate (usually not in this codebase)

### Audio Stream Handling
- Always initialize speaker with proper sample rate
- Lock speaker operations when modifying volume/state
- Handle stream errors gracefully
