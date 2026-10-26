
## Project conventions

- Use the Factory Pattern for constructing application components. Keep
  component creation in explicit factory types rather than scattering setup
  across handlers or the entry point.
- Keep every Go source and test file at or below 100 lines. Split files along
  cohesive responsibility boundaries when they approach the limit.
- Write README files, documentation, code comments, and generated reports in
  English.
- Preserve existing behavior unless a specification explicitly requires a
  behavior change. Update directly related documentation and tests.
