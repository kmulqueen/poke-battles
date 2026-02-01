# Frontend Repo Context

This skill defines how the frontend codebase is organized.

## Language & Tooling

- React 19
- TypeScript 5.9
- Vite
- CSS via modern layout primitives (Flexbox, Grid)

## Directory Structure

- `src/`
  - `components/`
    - Reusable presentational UI components
  - `pages/`
    - Route-level components
  - `hooks/`
    - Reusable React hooks
    - No direct DOM manipulation
  - `services/`
    - API and WebSocket clients
  - `state/`
    - Client-side state management
  - `utils/`
    - Pure helper functions
  - `types/`
    - Shared TypeScript types

## Ownership & Boundaries

- Components focus on rendering and user interaction
- Hooks encapsulate logic and side effects
- Services handle network communication
- UI must not embed backend assumptions

## Naming Conventions

- Components: `PascalCase.tsx`
- Hooks: `useSomething.ts`
- Utilities: `camelCase.ts`
- Files colocated when tightly coupled

## Responsibilities

- Semantic HTML is preferred
- Accessibility is a baseline expectation
- State and effects are explicit and traceable
