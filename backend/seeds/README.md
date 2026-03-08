# Seed Strategy

Seed data lives in environment-specific folders under this directory.

- `development/`: local developer fixtures and repeatable sample data
- `test/`: deterministic fixtures for test environments

Rules:

- seed files must be plain `.sql` files
- files run in lexical order
- keep seeds idempotent whenever possible
- schema changes belong in `backend/migrations`, not here
