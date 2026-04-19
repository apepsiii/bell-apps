# Database Migrations

This project supports both **SQLite** (default) and **MySQL/MariaDB**.

## Quick Start

### SQLite (Default)
```bash
./bell
# Database will be created automatically at ./database.db
```

### MySQL/MariaDB
```bash
export DB_DRIVER=mysql
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your_password
export DB_NAME=bell

./bell
```

Or with docker:
```bash
docker run -d \
  -e DB_DRIVER=mysql \
  -e DB_HOST=mysql-container \
  -e DB_PORT=3306 \
  -e DB_USER=bell \
  -e DB_PASSWORD=secret \
  -e DB_NAME=bell \
  bell:latest
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_DRIVER` | `sqlite` | `sqlite` or `mysql` |
| `DB_HOST` | `localhost` | MySQL host |
| `DB_PORT` | `3306` | MySQL port |
| `DB_USER` | `root` | MySQL user |
| `DB_PASSWORD` | `` | MySQL password |
| `DB_NAME` | `bell` | Database name |

## Manual Migration with golang-migrate

If you want to use `golang-migrate` CLI:

```bash
# Install migrate CLI
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -database "mysql://user:pass@tcp(localhost:3306)/bell" -path migrations up

# Rollback
migrate -database "mysql://user:pass@tcp(localhost:3306)/bell" -path migrations down
```

## Migration Files

| File | Description |
|------|-------------|
| `000001_init_schema.up.sql` | Create all tables and seed data |
| `000001_init_schema.down.sql` | Drop all tables |

## MySQL Schema Reference

```sql
-- Create database
CREATE DATABASE bell CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create user (optional)
CREATE USER 'bell'@'localhost' IDENTIFIED BY 'secret';
GRANT ALL PRIVILEGES ON bell.* TO 'bell'@'localhost';
FLUSH PRIVILEGES;
```
