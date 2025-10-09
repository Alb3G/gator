# Gator - RSS Feed Aggregator CLI

Gator is an RSS feed aggregator built in Go that allows users to follow, manage, and read posts from their favorite RSS feeds via the command line.

## Features

- **User management**: Register, login, and manage multiple users
- **Feed aggregation**: Add and manage RSS feeds from any source
- **Follow system**: Follow/unfollow feeds individually
- **Automatic collection**: Periodically fetch new posts from feeds
- **Post browsing**: View the latest posts from your followed feeds
- **PostgreSQL database**: Persistent storage for users, feeds, and posts
- **SQLC code generation**: Type-safe SQL queries automatically generated

## Requirements

- **Go**: 1.24.3 or higher
- **PostgreSQL**: Running PostgreSQL database
- **SQLC**: To generate Go code from SQL (optional, development only)
- **Goose**: For database migrations (optional, development only)

## Installation

### 1. Clone the repository

```bash
git clone https://github.com/Alb3G/gator.git
cd gator
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Set up PostgreSQL

Make sure you have PostgreSQL installed and running. Create a database for Gator:

```bash
createdb gator
```

### 4. Apply database migrations

If you have Goose installed:

```bash
cd sql/schema
goose postgres "postgres://username:password@localhost:5432/gator" up
```

Or manually apply the SQL files in numerical order from [sql/schema](sql/schema/).

### 5. Configure the configuration file

Create the `.gatorconfig.json` file in your home directory (`~/.gatorconfig.json`):

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Replace `username`, `password`, and database name according to your PostgreSQL setup.

### 6. Build the project

```bash
go build -o gator
```

Optionally, move the binary to your PATH:

```bash
sudo mv gator /usr/local/bin/
```

## Usage

### Available Commands

#### `register <username>`

Register a new user and set them as the current user.

```bash
gator register my_user
```

#### `login <username>`

Switch to the specified user.

```bash
gator login my_user
```

#### `users`

List all registered users, marking the current user with an asterisk.

```bash
gator users
```

**Example output:**
```
* my_user (current)
* another_user
```

#### `addfeed <name> <url>`

Add a new RSS feed and automatically follow it. **Requires being logged in.**

```bash
gator addfeed "Hacker News" https://news.ycombinator.com/rss
gator addfeed "Go Blog" https://go.dev/blog/feed.atom
```

#### `feeds`

List all available feeds with their owner user. **Requires being logged in.**

```bash
gator feeds
```

#### `follow <url>`

Follow an existing feed. **Requires being logged in.**

```bash
gator follow https://news.ycombinator.com/rss
```

#### `following`

Show all feeds the current user is following. **Requires being logged in.**

```bash
gator following
```

#### `unfollow <url>`

Unfollow a feed. **Requires being logged in.**

```bash
gator unfollow https://news.ycombinator.com/rss
```

#### `agg <interval>`

Start the aggregator that collects new posts from feeds periodically. The interval must be in Go duration format (e.g., `1m`, `30s`, `1h`).

```bash
gator agg 1m    # Fetch feeds every minute
gator agg 30s   # Fetch feeds every 30 seconds
gator agg 1h    # Fetch feeds every hour
```

**Note:** This command runs continuously. Press `Ctrl+C` to stop it.

#### `browse [limit]`

Display the latest posts from the feeds you follow. Optionally specify a limit (default: 2).

```bash
gator browse       # Show 2 posts
gator browse 10    # Show 10 posts
```

#### `reset`

**⚠️ WARNING:** Deletes all data from the database (users, feeds, follows, and posts).

```bash
gator reset
```

## Project Architecture

```
gator/
├── main.go                          # Application entry point
├── internal/
│   ├── commands.go                  # Implementation of all commands
│   ├── config/
│   │   └── config.go               # Configuration and state management
│   ├── database/                    # SQLC generated code
│   │   ├── db.go
│   │   ├── models.go
│   │   ├── users.sql.go
│   │   ├── feeds.sql.go
│   │   ├── feed_follows.sql.go
│   │   └── posts.sql.go
│   ├── rss/
│   │   └── rss.go                  # RSS client for fetching feeds
│   └── utils/
│       └── utils.go                 # Helper functions
├── sql/
│   ├── schema/                      # Database migrations
│   │   ├── 001_users.sql
│   │   ├── 002_feeds.sql
│   │   ├── 003_feed_follow.sql
│   │   ├── 004_add_last_fetched_to_feeds.sql
│   │   └── 005_posts.sql
│   └── queries/                     # SQL queries for SQLC
│       ├── users.sql
│       ├── feeds.sql
│       ├── feed_follows.sql
│       └── posts.sql
├── sqlc.yaml                        # SQLC configuration
├── go.mod
└── go.sum
```

### Main Components

- **Commands**: CLI command registration and execution system
- **Middleware**: Logged-in user validation for protected commands
- **State**: Maintains application state (configuration and DB queries)
- **RSS Client**: Parses RSS feeds using XML
- **SQLC**: Generates type-safe Go code from SQL queries

## Data Model

### Tables

#### `users`
- `id`: UUID (PK)
- `created_at`: TIMESTAMP
- `updated_at`: TIMESTAMP
- `user_name`: TEXT UNIQUE

#### `feeds`
- `id`: UUID (PK)
- `created_at`: TIMESTAMP
- `updated_at`: TIMESTAMP
- `name`: TEXT
- `url`: TEXT UNIQUE
- `user_id`: UUID (FK → users)
- `last_fetched_at`: TIMESTAMP (nullable)

#### `feed_follows`
- `id`: UUID (PK)
- `created_at`: TIMESTAMP
- `updated_at`: TIMESTAMP
- `user_id`: UUID (FK → users)
- `feed_id`: UUID (FK → feeds)

#### `posts`
- `id`: UUID (PK)
- `created_at`: TIMESTAMP
- `updated_at`: TIMESTAMP
- `title`: TEXT
- `url`: TEXT UNIQUE
- `description`: TEXT (nullable)
- `published_at`: TIMESTAMP
- `feed_id`: UUID (FK → feeds)

## Typical Workflow

1. **Register and login:**
   ```bash
   gator register john
   ```

2. **Add some feeds:**
   ```bash
   gator addfeed "TechCrunch" https://techcrunch.com/feed/
   gator addfeed "The Verge" https://www.theverge.com/rss/index.xml
   ```

3. **Start the aggregator in the background:**
   ```bash
   gator agg 5m &
   ```
   Or in another terminal:
   ```bash
   gator agg 5m
   ```

4. **Browse posts:**
   ```bash
   gator browse 20
   ```

5. **Follow feeds from other users:**
   ```bash
   gator feeds          # View all available feeds
   gator follow <url>   # Follow a specific feed
   ```

6. **Manage followings:**
   ```bash
   gator following      # View feeds you follow
   gator unfollow <url> # Unfollow a feed
   ```

## Development

### Generate database code

If you modify SQL queries in [sql/queries/](sql/queries/), regenerate the Go code:

```bash
sqlc generate
```

### Create new migrations

```bash
cd sql/schema
goose create migration_name sql
```

### Apply migrations

```bash
goose postgres "your_connection_string" up
```

### Rollback migrations

```bash
goose postgres "your_connection_string" down
```

## Dependencies

- [github.com/lib/pq](https://github.com/lib/pq) - PostgreSQL driver for Go
- [github.com/google/uuid](https://github.com/google/uuid) - UUID generation

## Technologies Used

- **Go 1.24.3**: Main programming language
- **PostgreSQL**: Relational database
- **SQLC**: Go code generator from SQL
- **Goose**: Database migration tool
- **RSS/XML**: RSS feed parsing

## Contributing

Contributions are welcome. Please:

1. Fork the project
2. Create a feature branch (`git checkout -b feature/new-feature`)
3. Commit your changes (`git commit -m 'Add new feature'`)
4. Push to the branch (`git push origin feature/new-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

## Author

Alberto Guzmán ([Alb3G](https://github.com/Alb3G))

## Support

If you encounter any issues or have suggestions, please open an [issue](https://github.com/Alb3G/gator/issues) on GitHub.

---

Enjoy aggregating and reading your favorite RSS feeds with Gator! 📰
