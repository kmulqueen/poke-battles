# Frontend Design System (Tailwind CSS)

This skill defines frontend design and styling standards using Tailwind CSS v4.1.

## Core Principles

- Mobile-first by default
- Responsive at all breakpoints
- Utility-first with intentional abstraction
- Design consistency through theme variables
- Accessibility-aware styling

## Responsive Design

- Base styles target the smallest supported viewport (320x560)
- Use Tailwind's mobile-first breakpoints:
  - Unprefixed utilities = mobile
  - sm, md, lg, xl, 2xl progressively enhance
- Avoid desktop-first overrides
- Layouts must adapt cleanly from phones → tablets → large desktops

Reference:

- Tailwind Responsive Design documentation

## Tailwind Usage Guidelines

### Utilities

- Prefer utilities over custom CSS
- Avoid arbitrary values unless necessary
- Group related utilities logically (layout → spacing → typography → visuals)

### Theme Variables (@theme)

- Use `@theme` for:
  - Colors
  - Spacing scales
  - Font families
  - Border radius
- Do not hardcode design tokens in components

### Base Styles (@base)

- Use for:
  - Element-level defaults (body, headings, buttons)
  - Typography normalization
- Do not add layout or component styling here

### Component Classes (@components)

- Use for:
  - Reusable UI patterns (buttons, cards, inputs)
  - Semantic abstractions over repeated utility sets
- Components should remain composable and override-friendly

### Variants (@variants)

- Use variants to express:
  - Interaction states (hover, focus, active)
  - Contextual states (disabled, selected)
- Prefer variants over duplicated utility blocks

## Accessibility & Semantics

- Styling should reinforce semantic HTML
- Visual hierarchy must match document structure
- Ensure sufficient contrast and focus visibility

## Anti-Patterns

- Desktop-first design
- Fixed widths that break responsiveness
- Deeply nested custom CSS
- Component-specific magic numbers
