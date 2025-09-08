# Contributing to Gitallica

We welcome contributions to Gitallica! This guide will help you get started with contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Process](#contributing-process)
- [Code Standards](#code-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Release Process](#release-process)

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Basic understanding of Git concepts and software engineering metrics

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/yourusername/gitallica.git
   cd gitallica
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/bgricker/gitallica.git
   ```

## Development Setup

### Build the Project

```bash
go build -o gitallica .
```

### Run Tests

```bash
go test ./cmd
go test ./...
```

### Test Commands

```bash
# Test individual commands
./gitallica churn --help
./gitallica survival --last 30d
./gitallica bus-factor --path src/
```

## Contributing Process

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

Follow the [Code Standards](#code-standards) and [Testing](#testing) guidelines.

### 3. Test Your Changes

```bash
go test ./cmd
go test ./...
go build -o gitallica .
./gitallica --help
```

### 4. Commit Your Changes

Use conventional commit messages:

```bash
git commit -m "feat: add new metric for code complexity analysis"
git commit -m "fix: correct threshold calculation in bus-factor command"
git commit -m "docs: update user guide with new examples"
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Code Standards

### Go Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` and `golint` for formatting
- Write clear, self-documenting code
- Use meaningful variable and function names

### Project Structure

```
cmd/
├── command_name.go          # Command implementation
├── command_name_test.go     # Command tests
└── utils.go                # Shared utilities

docs/
├── USER_GUIDE.md           # User documentation
├── COMMANDS.md             # Command reference
└── RESEARCH.md             # Research methodology

README.md                   # Project overview
CONTRIBUTING.md            # This file
LICENSE                    # MIT License
```

### Command Implementation Pattern

Each command should follow this pattern:

```go
// commandCmd represents the command
var commandCmd = &cobra.Command{
    Use:   "command-name",
    Short: "Brief description",
    Long: `Detailed description with examples and research basis.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation
    },
}

func init() {
    rootCmd.AddCommand(commandCmd)
    // Add flags
}
```

### Error Handling

- Use `errors.New()` for sentinel errors
- Provide clear, actionable error messages
- Handle edge cases gracefully
- Log errors appropriately

### Documentation

- Document all public functions
- Include research basis for metrics
- Provide usage examples
- Update README and docs for new features

## Testing

### Test-Driven Development

We follow TDD principles:

1. Write tests first
2. Implement functionality
3. Refactor and optimize
4. Ensure all tests pass

### Test Structure

```go
func TestCommandName(t *testing.T) {
    // Test basic functionality
}

func TestCommandNameEdgeCases(t *testing.T) {
    // Test edge cases
}

func TestCommandNameIntegration(t *testing.T) {
    // Test integration scenarios
}
```

### Test Requirements

- All new code must have tests
- Test coverage should be >90%
- Include edge cases and error conditions
- Test both success and failure scenarios

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./cmd

# Run tests with coverage
go test -cover ./cmd

# Run tests with verbose output
go test -v ./cmd
```

## Documentation

### README Updates

- Update feature list for new commands
- Add usage examples
- Update installation instructions
- Keep research citations current

### User Guide Updates

- Add new commands to [USER_GUIDE.md](docs/USER_GUIDE.md)
- Include examples and use cases
- Update troubleshooting section

### Command Reference Updates

- Add new commands to [COMMANDS.md](docs/COMMANDS.md)
- Document all flags and options
- Provide examples for each command

### Research Documentation

- Update [RESEARCH.md](docs/RESEARCH.md) for new metrics
- Include research citations
- Document threshold methodology
- Explain statistical analysis

## Types of Contributions

### New Metrics

When adding new metrics:

1. **Research Foundation**: Provide authoritative source
2. **Threshold Definition**: Define clear thresholds
3. **Implementation**: Follow existing patterns
4. **Testing**: Comprehensive test coverage
5. **Documentation**: Update all relevant docs

### Bug Fixes

- Identify the issue clearly
- Write tests that reproduce the bug
- Implement the fix
- Ensure all tests pass
- Update documentation if needed

### Performance Improvements

- Measure current performance
- Implement optimization
- Verify improvement
- Ensure no regressions
- Update documentation

### Documentation Improvements

- Fix typos and grammar
- Improve clarity and examples
- Add missing information
- Update outdated content

## Release Process

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] Version number updated
- [ ] CHANGELOG.md updated
- [ ] Release notes prepared
- [ ] Tag created on GitHub

### Creating a Release

1. Update version in `main.go`
2. Update `CHANGELOG.md`
3. Commit changes
4. Create and push tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
5. Create release on GitHub

## Community Guidelines

### Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Help others learn and grow
- Follow the golden rule

### Communication

- Use clear, descriptive commit messages
- Provide context in pull requests
- Ask questions when unsure
- Share knowledge and insights

### Recognition

Contributors will be recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project documentation

## Getting Help

### Questions and Discussion

- Open an issue for questions
- Use GitHub Discussions for general discussion
- Check existing issues and PRs first

### Development Help

- Review existing code for patterns
- Check test files for examples
- Read documentation thoroughly
- Ask for clarification when needed

## License

By contributing to Gitallica, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Gitallica! Your contributions help make software engineering more data-driven and effective. 🎸
