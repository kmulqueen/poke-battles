# Frontend Testing Standards

This skill defines frontend testing expectations.

## Principles

- Test user-visible behavior, not implementation
- Prefer RTL queries that reflect accessibility
- Avoid brittle snapshot tests

## Preferred Testing Style

- Interact with components via user events
- Assert on rendered output and ARIA roles
- Mock network boundaries, not UI internals
