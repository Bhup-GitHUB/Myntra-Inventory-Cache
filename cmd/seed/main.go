package main

import (
	"context"
	"log"
	"time"

	"inventory-cache-lab/internal/config"
	"inventory-cache-lab/internal/db"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := db.OpenMySQL(ctx, cfg.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := db.RunMigrations(conn, "migrations"); err != nil {
		log.Fatal(err)
	}
	if err := db.ResetDemoData(ctx, conn); err != nil {
		log.Fatal(err)
	}
	log.Println("demo data reset")
}
