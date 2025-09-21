# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based Spotify API client with CLI interface. The goal is to create a comprehensive client that enables access to the full functionality of the Spotify API.

**Components:**
- **CLI**: Golang-based command line interface
- **API Client**: Golang library for Spotify API integration

## Development Setup

This project uses Go as the primary language for both the API client and CLI components.

## Common Commands

### For Go projects:
- `go mod init` - Initialize Go module
- `go mod tidy` - Clean up dependencies
- `go build` - Build the project
- `go test ./...` - Run all tests
- `go run main.go` - Run the application
- `gofmt -w .` - Format Go code
- `go vet ./...` - Run Go vet for static analysis

## Development Lifecycle

1. Create feature branch based off ticket that is being worked on
2. Create a plan for the task and present it
3. Wait for approval before getting started
4. After each step in the implementation, commit to the feature branch following conventional commits
5. Once the work is done, push the branch up and setup a review
6. Once the branch is merged, make sure the issue is closed

## Spotify API Integration

When working with Spotify API:

- Store credentials securely using environment variables
- Use the Spotify Web API SDK when available
- Implement proper OAuth 2.0 flow for user authentication
- Handle rate limiting (429 responses) appropriately
- Use refresh tokens for long-lived sessions

## Architecture Considerations

For this Go-based Spotify API project, consider:

- **Authentication**: OAuth 2.0 flow implementation
- **Data Models**: Track, Album, Artist, Playlist entities (Go structs)
- **Caching**: Implement caching for API responses to reduce rate limiting
- **Error Handling**: Robust error handling for API failures and rate limits
- **CLI Design**: User-friendly command structure and output formatting
- **Package Structure**: Separate packages for API client and CLI components

## Project Roles

### Project Manager
- Break down the project into milestones with associated tasks
- Each task should have description, acceptance criteria, and testing requirements

### Backend Engineer
- Implementation of tasks outlined by the Project Manager

## Environment Variables

Typical environment variables needed:
- `SPOTIFY_CLIENT_ID`
- `SPOTIFY_CLIENT_SECRET`
- `SPOTIFY_REDIRECT_URI`

## Testing

When implementing tests:
- Mock Spotify API responses for unit tests
- Test authentication flows
- Test error handling scenarios
- Consider integration tests with actual API (using test credentials)

## Lessons Learned

### Milestone 1: Project Foundation & Core Infrastructure

**Standard Library First Approach:**
- Prefer Go standard library over external dependencies when possible
- `errors` package with Go 1.13+ wrapping is sufficient for typed errors
- Standard `log` package can be extended for structured logging
- Reduces dependency bloat and improves maintainability

**Configuration Management:**
- Multi-source configuration (files → .env → env vars) provides excellent flexibility
- Validation at load time prevents runtime configuration errors
- Default values ensure the application can run with minimal setup
- Sensitive data should always come from environment variables, never committed files

**Development Workflow:**
- Atomic commits with conventional commit messages improve project history
- Feature branches with descriptive names aid in tracking work
- Comprehensive PR descriptions with test plans improve code review quality
- Documentation updates should be committed separately from code changes

**Project Structure:**
- Follow Go project layout standards: `cmd/`, `pkg/`, `internal/`
- Internal packages prevent accidental imports and provide clear boundaries
- Comprehensive `.gitignore` prevents accidental commits of sensitive data
- Clear separation of concerns between config, logging, and error handling

**Testing Strategy:**
- Unit tests for each package verify functionality in isolation
- Test error paths as thoroughly as happy paths
- Verbose test output (`go test -v`) helps verify test coverage
- Build verification (`go build ./...`) ensures all packages compile correctly

### Milestone 2: Authentication & Authorization

**OAuth 2.0 Implementation:**
- Research API documentation thoroughly before implementation - initial assumption about simple API keys was incorrect
- Spotify requires proper OAuth 2.0 flows (Client Credentials + Authorization Code)
- Client Credentials flow is perfect for CLI tools accessing public data
- Authorization Code flow enables user-specific data access (playlists, library)
- Always implement token refresh mechanism for long-running applications

**HTTP Client Design:**
- Automatic token management reduces complexity for API consumers
- Context-aware requests provide proper cancellation and timeout handling
- Centralized error handling with typed errors improves debugging
- Pre-flight token validation prevents unnecessary API calls

**Security Considerations:**
- Base64 encoding of client credentials follows OAuth 2.0 standards
- Token expiration checks prevent API failures
- Error messages should not leak sensitive credential information
- Proper timeout handling prevents hanging requests

**Testing OAuth Flows:**
- Unit tests should focus on logic, not external API calls
- Integration tests require real credentials and should be optional
- Mock servers are useful for testing HTTP request structure
- Skip tests that require external dependencies rather than failing

**API Client Architecture:**
- Separation of concerns: auth package handles OAuth, client package handles HTTP
- Dependency injection would improve testability for future enhancements
- Consistent error types across packages improve error handling
- Token management should be transparent to API endpoint implementations