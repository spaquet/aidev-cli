# Documentation

This folder contains comprehensive documentation for the AIDev CLI project. Below is a guide to each document and when to use it.

## For Users

Start here if you're using the AIDev CLI:

- **[Main README](../README.md)** — Installation, quick start, commands, configuration, SSH setup, troubleshooting
  - Read this first for getting started
  - Contains all common use cases and FAQs

## For Developers

- **[Architecture](ARCHITECTURE.md)** — System design and implementation details
  - Read this to understand how the project is structured
  - Describes components, TUI architecture, authentication flow, and design decisions

- **[Contributing](CONTRIBUTING.md)** — Development workflow, CI/CD pipeline, testing
  - How to set up your development environment
  - How to run tests locally
  - How to create a release

## API Documentation

- **[API Reference](rails-api-spec.md)** — Backend REST API specification
  - All endpoints, request/response formats
  - Authentication headers
  - SSE event types for real-time updates

- **[Authentication Spec](auth-spec.md)** — JWT token lifecycle and storage
  - How tokens are issued and refreshed
  - Config file format and management
  - Implementation details for backend teams

## UI/UX Documentation

- **[TUI Design](tui-design.md)** — User interface specification
  - All screens and their layouts
  - Navigation flow and keybindings
  - Visual style and color palette
  - Modal dialogs and interactions

## Document Map

```
docs/
├── README.md                  ← You are here
├── ARCHITECTURE.md            ← How the system works
├── CONTRIBUTING.md            ← How to develop & release
├── tui-design.md              ← UI specification
├── rails-api-spec.md          ← Backend API spec
└── auth-spec.md               ← Authentication details
```

## Quick Links

- **Installation**: See main [README](../README.md#installation)
- **Command Reference**: See main [README](../README.md#commands)
- **Configuration**: See main [README](../README.md#configuration)
- **Troubleshooting**: See main [README](../README.md#troubleshooting)

## Contributing

If you're contributing code:

1. Start with [Architecture](ARCHITECTURE.md) to understand the codebase
2. Read [Contributing](CONTRIBUTING.md) for the development workflow
3. Check [TUI Design](tui-design.md) if modifying the interface
4. Refer to [API Reference](rails-api-spec.md) if integrating with the backend

## Keeping Documentation Current

- Documentation is expected to stay in sync with the code
- When implementing a feature, update the relevant doc files
- When changing the API, update `rails-api-spec.md`
- When changing the UI, update `tui-design.md`
- When changing architecture, update `ARCHITECTURE.md`

## Questions?

- Check the [main README](../README.md) for user questions
- Open an issue on GitHub for documentation gaps
- See [Contributing](CONTRIBUTING.md) for development questions
