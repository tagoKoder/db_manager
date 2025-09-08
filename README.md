# command_service
This project is a CLI tool for managing database-related operations, such as generating schemas and applying migrations. It helps centralize processes involving multiple modules or services.

---

# Database Migration Commands (migrate)
This CLI service is designed to manage database migrations in a controlled and flexible way. It uses the cobra framework and allows you to apply or revert migrations from the command line.

## Command Structure

command migrate [up|down] [flags]
The migrate command contains two main subcommands:

up: Applies migrations.

down: Reverts migrations.

Both commands require specifying the .env file and accept additional parameters.

🚀 Subcommand up
Applies migrations to the database. You can apply all or a specific number of steps/versions.

```
go run cmd/main.go command migrate up --env=env/.env.<file> [--all]
```
Examples

command migrate up --env .env.local --all

🔁 Subcommand down
Reverts applied migrations. You can revert all, or only a specific number of steps or files.

```
go run cmd/main.go command migrate down --env <path> [--all] [--steps-version N] [--steps-file N]
```
Examples

command migrate down --env .env.local --all
command migrate down --env .env.dev --steps-version 1
command migrate down --env .env.staging --steps-file 2