# Major Refactoring Plan

## Overview
This branch is dedicated to major refactoring of the Pi-hole Network Analyzer codebase to improve maintainability, testability, and code organization.

## Primary Targets

### 1. Extract Functions from main.go (1693 lines)
**Priority: HIGH**

The monolithic `cmd/pihole-analyzer/main.go` file needs to be broken down into logical packages:

#### Planned Extractions:
- **`internal/analyzer/`** - Core DNS analysis logic
- **`internal/ssh/`** - SSH connection and Pi-hole database operations  
- **`internal/csv/`** - CSV parsing and processing
- **`internal/network/`** - ARP table analysis and network utilities
- **`internal/reporting/`** - Report generation and file output
- **`internal/cli/`** - Command-line interface and flag handling

### 2. Resolve Naming Inconsistency
**Priority: HIGH**

Current inconsistent naming:
- Module: `pihole-network-analyzer`
- Makefile binary: `dns-analyzer`
- Command directory: `cmd/pihole-analyzer/`

**Decision**: Standardize on `pihole-analyzer` throughout the project.

### 3. Improve Error Handling
**Priority: MEDIUM**

- Implement structured error handling
- Add proper error wrapping and context
- Improve SSH connection error recovery
- Add retry mechanisms for network operations

### 4. Add Structured Logging
**Priority: MEDIUM**

Replace `fmt.Printf` statements with structured logging:
- Use `log/slog` or similar structured logging library
- Add log levels (DEBUG, INFO, WARN, ERROR)
- Improve debugging capabilities

### 5. Configuration Validation
**Priority: MEDIUM**

- Add comprehensive configuration validation
- Improve error messages for invalid configurations
- Add configuration schema documentation

## Implementation Strategy

### Phase 1: Extract Core Functions
1. Create new internal packages
2. Move related functions from main.go
3. Maintain API compatibility
4. Add unit tests for extracted packages

### Phase 2: Improve Architecture
1. Implement dependency injection patterns
2. Add interfaces for better testability
3. Improve separation of concerns

### Phase 3: Polish and Optimize
1. Add structured logging
2. Improve error handling
3. Add configuration validation
4. Performance optimizations

## Testing Strategy

- Run full test suite after each extraction
- Maintain integration test compatibility
- Add unit tests for new packages
- Ensure colorized output still works correctly

## Success Criteria

- [ ] main.go reduced to <500 lines (entry point only)
- [ ] All functions properly organized in internal packages
- [ ] Naming consistency resolved throughout project
- [ ] All tests passing
- [ ] No regression in functionality
- [ ] Improved code maintainability scores

## Notes

- Follow the guidance in `.github/copilot-instructions.md`
- Preserve the project's core values: beautiful terminal output, robust SSH connectivity, comprehensive DNS analysis
- Maintain backward compatibility for CLI interface
- Keep integration with existing build system (Makefile)
