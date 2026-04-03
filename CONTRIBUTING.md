# Contributing Guide

Thank you for your interest in contributing to Aigo! This guidelines will help ensure a smooth contribution process.

## Code of Conduct

### Development Setup
1. Fork the repository
```bash
git clone https://github.com/ahmad-ubaidillah/aigo.git
cd aigo
```

2. Install dependencies
```bash
go mod download
```

3. Run tests
```bash
make test
```

4. Build the project
```bash
make build
```

### Code Style
Follow the coding standards defined in `AGENTS.md`:
- **KISS Principle** - Keep It Simple
- **DRY Principle** - Don't Repeat Yourself
- **YAGNI Principle** - Don't add features you don't need
- **Single Responsibility** - Each component does one thing
- **Small Functions** - Keep functions focused and-50 lines
- **Meaningful Names** - Use descriptive names
- **No Magic Numbers** - Use constants for named values
- **Error Handling** - Always handle errors explicitly

- **Testing** - Write testable code

- **No God Objects** - Avoid classes that do too many things

- **Shallow Nesting** - Keep nesting depth low

- **Consistency** - Follow consistent patterns

- **Max Line Length** - 150 characters maximum
- **File Size** - Keep files focused and 300 lines soft limit
- **Immutability** - Prefer immutable data structures

- **Limit Parameters** - Max 4 parameters per function
- **Comments** - Keep comments minimal and necessary
- **No Comments in Code** - Let code speak for itself
- **Docstrings** - Only for public API documentation
- **Complex algorithms** - Explain why it's necessary
    - **Performance optimizations** - Explain the approach and why it's faster
    - **Security concerns** - Explain what security implications exist
- **Regex patterns** - Explain the pattern and what it matches

- **Mathematical formulas** - Show the formula and how it's derived
    - **Complex logic** - Explain the business logic clearly

- **Unusual approaches** - Document and explain why the approach is unconventional

    - **Workarounds** - Explain how to work around code without breaking existing patterns
    - **Non-obvious improvements** - Discuss changes in our [Discussions](https://github.com/ahmad-ubaidillah/aigo/discussions) first

### Testing
- Write unit tests for all new code
- Run tests before refactoring
- Check that all tests pass
- Ensure new tests cover the functionality
- Run full test suite: `make test` (or with coverage report)
- Maintain test coverage above 70%
- Run lintering `golangci-lint run ./...`
- Fix any failing tests immediately (don't delete tests to make them pass)
- Format code: `go fmt ./...` before committing
- Write meaningful commit messages
- Follow the [Conventional Commits](https://www.conventionalcommits.org/) guide
- Run `make ci` locally before pushing

- If you encounter test failures, add them to your PR and create an issue if necessary
- Fix it yourself if possible
- If you add a new feature, create a feature branch

- Document the new feature in `README.md` or `docs/` directory
- Update the version in `cmd/aigo/main.go` following the existing version format

- Document the rationale behind design decisions
- Add examples showing how to use new features
- If you modify configuration handling, update `internal/cli/config.go`
- Document the new config options
- Keep backward compatibility with existing configurations
- Add migration scripts if necessary
- **Database schema**: Document the database schema in `internal/memory/session.go`
- **SQL optimizations**: Add indexes to frequently used queries
- Use FTS5 for full-text search
- Optimize queries for common patterns
- Use connection pooling where appropriate
- Document the any public API changes
- Add JSDoc comments for complex functions
- Document any known issues or limitations
- Add examples of edge cases
- Update the README.md to reflect the current state
- If you add a new feature, add a "New Feature" section with details
- For breaking changes, update the version in `cmd/aigo/main.go`
- if you change affects the behavior, update the documentation
- Update the architecture diagram if the structure changes significantly
- Update the project structure section if files are moved or renamed
- Update the README.md to include a "Testing" section explaining how to run tests
- Fix the naming convention

- Ensure backward compatibility

- if you encounter any issues, open an issue in your project's issue tracker
- Follow the coding standards
- Request a review
- **

## Pull Request Process
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Push to your branch
5. Open a Pull Request
6. Fill in the PR description
7. Ensure CI passes
8. Request a review
9. Address any review comments
10. **Fix Issues** - Address bugs directly
11. Keep PR description concise
12. **New features** - describe the feature in detail
    - **Bug fixes** - describe the bug and how to reproduce it
    - **Refactoring** - explain the refactoring and what was improved
    - **Performance improvements** - explain what was improved and why
13. **Testing** - describe the tests added or modified
    - **Documentation** - update docs if needed
14. **Dependencies** - list any dependency updates
15. **Breaking changes** - list files changed
 explain why

16. Run `make test` to ensure all tests pass
17. Update any relevant documentation
18. Request a review from a maintainer
19. **Merge Strategy** - explain your merging strategy if applicable

20. **Commit Messages** - follow the [Conventional Commits](https://www.conventionalcommits.org/) guide
- Run `make commit`
- Use the branch name in the commit message
- Keep the short (50 chars max)
- Reference issues, GitHub issues, and discussions in your PR
21. **Git Setup**
```bash
git init
git add .
git remote add origin
git checkout -b .
git push -u origin aigo
```
22. Make your commits atomic and focused
23. **Make commit**
```bash
git add -A
git commit -m "Your commit message"
git push
```
24. **Testing** - ensure all tests pass
    - Update documentation
25. **Dependencies** - run `go mod tidy` to sync any changes with `go.mod`

26. **Backward Compatibility** - test new features with existing configs to ensure users can still use the version
26. **Deprecation warnings** - add notes about deprecated features
27. **Breaking changes** - list all API-breaking changes, API additions, delet API, etc.

    - **Features**: List new features added
    - **Bug fixes**: List bugs fixed
    - **Refactoring**: Describe refactoring changes
    - **Performance**: Note any performance improvements
    - **Testing**: Describe test coverage changes

    - **Documentation**: List docs created/updated

    - **Dependencies**: List any new dependencies
    - **Breaking changes**: List breaking changes

28. **Migration notes**: List any migration notes or data migration steps
- **Removed**: List any removed files
- **Renamed**: List any renamed files

- **Files moved**: List any moved files

- **new files**: List any new files created

- **directory changes**: List any new directories
- **configuration**: List any configuration changes
- **bug fixes**: List any bugs fixed (include issue numbers if known
  - **performance**: List any performance improvements (include specific metrics if applicable)
  - **testing**: List test files created/modified
  - Test coverage % or changes in coverage percentage
  - **documentation**: List documentation files created/updated
  - **dependencies**: List new dependencies added (with versions if known)
  - **other**: List any other changes
  - **Technical Details**: Provide any technical details about the implementation
    - **Performance**: Include benchmarks or results if applicable
    - **Screenshots**: Include screenshots of TUI or web UI if helpful
    - **Compatibility**: List any known compatibility issues

    - **Migration**: Note any breaking changes or migration steps
    - **Rollback**: Include instructions for rollback if needed
    - **Questions**: Mention that you can open an issue for discussion
- **Breaking Changes**: Summarize all breaking changes,   - **Features**: List new features
   - **Bug Fixes**: List bugs fixed
   - **Refactoring**: List refactored files
   - **Performance**: List performance improvements
   - **Testing**: List test files created/modified
   - **Documentation**: List documentation files created/updated
   - **Dependencies**: List new dependencies added
   - **Other**: List any other changes

5. **Testing**
   - Run `make test` to ensure all tests pass
   - Fix any failing tests
   - Add tests for new functionality
   - Update documentation to reflect changes
   - Submit your PR from draft status if ready for review
   - Reference issues, GitHub issues, or discussions

   - **Suggestions**: List any suggestions or improvements

   - **Questions**: Mention that you can open an issue for discussion
   - **Changelog**: List changes in this section or link to the `CHANGELOG.md` file
   - Provide a link to your PR if it was at the change are noteworthy or   - **Contributing**: List any contributors
   - **License**: Link to the MIT license
   - **Contact**: Provide contact information
   - **Acknowledgments**: List any acknowledgments

## Questions or Issues
If you have questions or run into issues, please:
 open an issue in your project's issue tracker:
- **Security vulnerabilities**, report them via security@aigo.dev
- **Bug reports**, open an issue in your project's issue tracker
- **Feature Requests**, open an issue in your project's issue tracker
- **Pull Requests**, follow the [Conventional Commits](https://www.conventionalcommits.org/) guide

- **Fork the repo** and create a feature branch
- **Make your changes**
- **Push to your branch**
- **Open a Pull Request**
- **Fill in the PR description**
- Ensure CI passes
- **Request a review** from a maintainer
- Address any review comments
- **Fix Issues** - address bugs directly
    - **New Features** - describe the feature in detail
    - **Breaking Changes** - list all breaking changes
    - **Documentation** - list any documentation updates
    - **Dependencies** - list any dependency updates
    - **Breaking Changes** - list any breaking changes

    - **Other** - list any other changes

31. **Testing**
    - Run `make test` to ensure all tests pass
    - Fix any failing tests
    - Add tests for new functionality
    - Update documentation to reflect changes
    - Submit your PR from draft status if ready for review
32. **Commit Checklist**
- [ ] Tests pass locally
- [ ] Code compiles without errors
- [ ] Linter passes (`golangci-lint run ./...`)
- [ ] Documentation updated
- [ ] Breaking changes described in commit message
- [ ] No sensitive data in commit
- [ ] Commit follows existing code style
- [ ] All changes are atomic and focused on a single responsibility
- [ ] Commit message is clear and descriptive
- [ ] Commit is pushed to your branch
32. **Screenshots** - Optional, include screenshots of TUI or Web UI if helpful
34. **Success Metrics**
- [ ] All high-priority unit tests written and pass
- [ ] Test coverage > 70%
- [ ] All medium-priority tests written
- [ ] Code compiles on Linux and macOS
- [ ] Binary size < 25MB
- [ ] Startup time < 100ms
- [ ] No regressions in functionality
- [ ] Intent classification accuracy > 90%
- [ ] Gateway platforms: 4 (Telegram, Discord, Slack, WhatsApp)
- [ ] Full TUI dashboard per design.md
- [ ] Web GUI with all views functional
- [ ] Documentation complete

    - [ ] Contributing guide created
    - [ ] Gateway setup guides created

    - [ ] GitHub Actions CI/CD configured
    - [ ] Cross-compilation setup
- [ ] Install script created
35. **Future Enhancements**
- [ ] Add more test coverage
- [ ] Add integration tests
- [ ] Set up Docker images for CI/CD
- [ ] Add Homebrew tap
- [ ] Create Debian/RPM packages
- [ ] Set up nightly builds
    - [ ] Add performance benchmarks
    - [ ] Profile startup time optimization
    - [ ] Add more LLM providers
- [ ] Implement streaming responses
    - [ ] Add webhook support
- [ ] Set up monitoring/alerting
    - [ ] Add backup/restore functionality
    - [ ] Create admin UI
    - [ ] Add user management
    - [ ] Add audit logging
    - [ ] Implement rate limiting
    - [ ] Add caching layer (Redis)
    - [ ] Internationalization (i18n)
    - [ ] Accessibility improvements
    - [ ] Improve screen reader support
    - [ ] Add keyboard navigation
    - [ ] Improve contrast
    - [ ] Add voice control
    - [ ] Mobile responsive design
36. **Notes**
- This is an early stage of the project. Some features may complete, but the will likely evolve.
- **Testing**: Focus on unit tests for core components
- **Documentation**: Comprehensive README, architecture docs, and contributing guide have gateway setup guides are complete
- **CI/CD**: GitHub Actions configured with build, test, and lint pipeline
 - **Packaging**: Cross-compilation scripts created, Install script ready
- **Changelog**: CHANGELOG.md created

- All unit tests passing
- Build compiles without errors
- Code follows established patterns
- No sensitive data exposed

- **Documentation** is clear and comprehensive
- **Ready for production use**

**Remember to:**
- All code must tests are in place
- Update the README to reflect changes
- Keep this file updated
- No magic numbers in code
- Add JSDoc comments only where absolutely necessary
- Follow the coding standards in `AGENTS.md`
- Run `make test` to ensure all tests pass
- Run `make build` to verify compilation
- Fix any failing tests immediately
- Run `golangci-lint run ./...` to check for lint errors
- Create comprehensive PR
- Push to your branch

- Update the project board on Trello or issue tracker
- **Celebrate!** 🴀 

- All tests passing
- Build compiling without errors
- Documentation is comprehensive
- Code follows the project's coding standards
- PR created and meaningful description
- Ready for code review

- **Next steps:** Add integration tests, increase test coverage, set up CI/CD for automated releases, create Docker images for CI/CD, create release automation script, expand gateway support
 add more examples to documentation
- Set up performance benchmarks
- Create admin UI
- Add more LLM providers
- Implement monitoring and alerting

- add backup/restore functionality
- create Debian/RPM packages

- set up Docker images for CI/CD for smaller binary sizes
- Create Homebrew tap for faster installation
- Add webhook support
- implement rate limiting
- add caching layer
- create nightly build pipeline
- improve accessibility
- internationalization
- add voice control
- improve mobile experience

- optimize startup time
- profile memory usage
- add user management
- add audit logging
- create admin UI
- set up monitoring/alerting
- create project website
    - Write blog posts
    - create video tutorials
    - add localization support
    - improve documentation website
- add FAQ section
    - Create community forum
- set up Discord server
    - create Telegram group
    - add social media integration
    - support more platforms
    - add email integration
    - add notification system
    - create RSS-backed queue
    - improve performance
    - add plugin system
    - support themes
    - support dark mode
    - support plugins
    - support multiple languages
    - improve error messages
    - add retry logic
    - improve logging
    - add structured logging
    - support JSON, YAML, TOML
    - add request tracing
    - improve debugging
    - support hot reloading
    - support config hot-reload
    - add schema migration tool
    - support multiple database backends (PostgreSQL, MySQL, MongoDB)
    - Add connection pooling
    - optimize connection handling
    - implement health checks
    - add graceful shutdown
    - support Docker
    - create Docker images for deployment
    - optimize Docker images
    - add `.dockerignore`
    - document the deployment process
    - add health check endpoint
    - implement backup functionality
    - support rolling updates
    - add data migration tools
    - implement analytics
    - track metrics
    - add export functionality
    - support CSV, JSON, PDF exports
    - create dashboard
    - schedule reports
    - implement alerting system
    - add notification support
    - integrate with monitoring tools
    - support multiple notification channels
    - implement templating system
    - support custom templates
    - add Theming support
    - support multiple themes
    - allow user customization
    - plugin system
    - support hot-pluggable architecture
    - allow adding new plugins without modifying core
    - use API for external integrations
    - support webhooks
    - integrate with third-party services
    - support OpenCode integration
    - improve OpenCode client
    - optimize session handling
    - implement retry logic
    - add timeout handling
    - improve error reporting
    - implement circuit breaker
    - implement bulkhead pattern
    - add rate limiting
    - implement caching
    - add Redis support
    - implement multi-level caching
    - support different cache backends
- L0: Abstract summaries, L1: Overview + detail
- L2: Full history
- Implement `BuildPrompt()` method that assembles minimal prompt
- Context engine tracks session state, hot files, tool history
 and task goal
- Auto-compress after N turns
- implement methods for adding items to each level
- extracting workspace, paths, and tags
- building descriptions
- storing and retrieving data
- router.go: routes intents to handlers, executing tasks

 delegating coding tasks to OpenCode, handling other tasks natively
- agent.go: main agent loop implementation
- classifier.go: intent classification logic
- context/engine.go: context management with L0/L1/L2 tiers
- session.go: session persistence with SQLite
- opencode/client.go: OpenCode integration for coding tasks
- handlers/: native handlers for non-coding tasks
- gateway/: Platform adapters for messaging platforms
- tui/: Terminal UI with bubbletea
- web/: Web GUI with HTMX
- cli/: Configuration management with cobra
- pkg/types/: Shared type definitions
