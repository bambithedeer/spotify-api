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