---
applyTo: "**/*.ts,**/*.tsx"
---
- Use strict TypeScript — no `any`, prefer `unknown` with type guards
- Use `interface` for object shapes, `type` for unions/intersections
- Prefer `readonly` for immutable properties
- Use named exports over default exports
- Handle errors with proper types, not string messages
