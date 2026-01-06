package main

import (
    "database/sql"
    "flag"
    "fmt"
    "log"
    "os"
    
    "github.com/pressly/goose/v3"
    _ "github.com/lib/pq"
    
    "inshorts-news-api/config"
    _ "inshorts-news-api/db/migrations" // Import to register migrations
)

var (
    flags = flag.NewFlagSet("migrate", flag.ExitOnError)
    dir   = flags.String("dir", "db/migrations", "directory with migration files")
)

func main() {
    flags.Parse(os.Args[1:])
    args := flags.Args()

    if len(args) < 1 {
        printUsage()
        os.Exit(1)
    }

    command := args[0]

    // Load config
    cfg := config.Load()
    
    // Build DSN
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
    )

    // Open database connection
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Set dialect
    if err := goose.SetDialect("postgres"); err != nil {
        log.Fatalf("Failed to set dialect: %v", err)
    }

    // Run goose command
    var cmdArgs []string
    if len(args) > 1 {
        cmdArgs = args[1:]
    }
    
    if err := goose.Run(command, db, *dir, cmdArgs...); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    log.Println("Migration completed successfully!")
}

func printUsage() {
    fmt.Println("Usage: migrate [OPTIONS] COMMAND")
    fmt.Println("\nCommands:")
    fmt.Println("  up                   Migrate to the most recent version")
    fmt.Println("  up-by-one            Migrate up by 1")
    fmt.Println("  up-to VERSION        Migrate to a specific version")
    fmt.Println("  down                 Roll back by 1")
    fmt.Println("  down-to VERSION      Roll back to a specific version")
    fmt.Println("  redo                 Re-run the latest migration")
    fmt.Println("  reset                Roll back all migrations")
    fmt.Println("  status               Show migration status")
    fmt.Println("  version              Print current migration version")
    fmt.Println("  create NAME [go|sql] Create a new migration file")
    fmt.Println("\nOptions:")
    flags.PrintDefaults()
}
