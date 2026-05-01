# bootdev_blogagg
Blog Aggregator guided project for Boot.Dev - Golang, Postgresql

Gator Setup Guide
1. Install prerequisites

Before running Gator, you’ll need:

Go

Install Go from:
https://go.dev/dl/

Verify:

go version
PostgreSQL

Install Postgres from:
https://www.postgresql.org/download/

After installation:

psql --version

Make sure your Postgres server is running, then create a database for Gator:

createdb gator
2. Install the Gator CLI

From the root of the project:

go install .

This builds and installs the gator binary into your Go bin directory.

If needed, make sure your Go bin path is in your shell:

export PATH=$PATH:$(go env GOPATH)/bin

Then verify:

gator
3. Set up your config file

Gator expects a config file in your home directory:

~/.gatorconfig.json

Example:

{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
Replace:
username → your Postgres username
password → your Postgres password
gator → your database name
4. Run database migrations

Before using Gator, make sure your database schema is set up.

(Depending on your migration tool)

Example:

goose postgres "postgres://username:password@localhost:5432/gator?sslmode=disable" up
5. Running Gator

Basic syntax:

gator <command> <args>
Useful commands
Register a user
gator register halim
Log in as a user
gator login halim
Add a feed
gator addfeed "Hacker News" "https://news.ycombinator.com/rss"
Follow a feed
gator follow "https://news.ycombinator.com/rss"
View followed feeds
gator following
Unfollow a feed
gator unfollow "https://news.ycombinator.com/rss"
Browse posts
gator browse
Reset database
gator reset
Typical first-time workflow
gator register yourname
gator login yourname
gator addfeed "Boot.dev Blog" "https://blog.boot.dev/index.xml"
gator follow "https://blog.boot.dev/index.xml"
gator browse
Common issues
“command not found: gator”

Your Go bin path probably isn’t in PATH.

Postgres connection errors

Double-check:

Username
Password
Database name
Postgres service running
Mental model

Gator is basically:

CLI (Go) → Config file → PostgreSQL DB → RSS feeds

Your CLI commands interact with Postgres, store users/feeds/follows, and fetch content from RSS sources.