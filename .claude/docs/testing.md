# Testing

## Running Tests

```bash
go test ./...                              # All tests
go test -v ./...                           # Verbose
go test ./api                              # Single package
go test -run TestBrowserImplGetStations ./api  # Single test
```

## Mock Pattern

Mocks use function fields for flexible behavior configuration:

```go
mockBrowser := &mocks.MockRadioBrowserService{
    GetStationsFunc: func(query StationQuery, ...) ([]Station, error) {
        return []Station{{Name: "Test Radio"}}, nil
    },
}
```

**Mock files:** `mocks/` directory
- `browser_mock.go` - RadioBrowserService
- `playback_manager_mock.go` - PlaybackManagerService
- `storage_mock.go` - StationStorageService
- `http_client_mock.go` - HTTPClientService

## Dependency Injection

```go
// Production
model, err := models.NewDefaultModel(cfg)

// Testing
model := models.NewModel(cfg, mockBrowser, mockPlayback, mockStorage)
```

## Test Organization

- Tests live alongside source: `foo.go` + `foo_test.go`
- Use `github.com/stretchr/testify/assert` for assertions
- Table-driven tests preferred:

```go
tests := []struct {
    name     string
    input    string
    expected string
}{
    {"empty", "", ""},
    {"basic", "test", "TEST"},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        assert.Equal(t, tt.expected, transform(tt.input))
    })
}
```

## Playback Testing

The playback package uses interface abstractions for exec.Command:

- `CommandExecutor` - wraps `exec.Command()`
- `Cmd` - wraps `exec.Cmd`
- `Process` - wraps `os.Process`

This allows testing FFplay integration without actually spawning processes.
