package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/NurulloMahmud/medicalka-project/config"
	"github.com/NurulloMahmud/medicalka-project/internal/platform/database"
	"github.com/NurulloMahmud/medicalka-project/internal/user"
)

func main() {
	cfg := config.Load()

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := user.NewPostgresRepository(db)

	deleted, err := repo.DeleteUnverifiedUsers(context.Background())
	if err != nil {
		log.Fatalf("cleanup failed: %v", err)
	}

	fmt.Fprintf(os.Stdout, "cleanup completed: deleted %d unverified users older than %d hours\n", deleted)
}
