# Testdata

This directory contains test fixtures and sample data for Kamui unit tests.

## Structure

- `sessions/` - Sample session JSON files for testing serialization and deserialization
- `fixtures/` - Mock Claude session files and other test fixtures

## Usage

Test files can reference these fixtures using relative paths:

```go
sessionData, err := os.ReadFile("../../testdata/sessions/example-session.json")
```

## Files

- `sessions/example-session.json` - Complete example session with all fields populated
- `fixtures/claude-session.jsonl` - Sample Claude session conversation in JSONL format

These files are used by the test suite to ensure proper handling of real-world data formats.