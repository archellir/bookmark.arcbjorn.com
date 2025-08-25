package main

import (
	"flag"
	"fmt"
	"log"

	"torimemo/internal/db"
	"torimemo/internal/models"
)

func main() {
	var dbPath string
	flag.StringVar(&dbPath, "db", "./torimemo.db", "Database path")
	flag.Parse()

	fmt.Println("🌱 Seeding demo data...")

	// Initialize database
	database, err := db.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize repository
	bookmarkRepo := db.NewBookmarkRepository(database)

	// Helper function to create string pointer
	stringPtr := func(s string) *string { return &s }

	// Demo bookmarks
	demoBookmarks := []models.CreateBookmarkRequest{
		{
			Title:       "GitHub - Go Programming Language",
			URL:         "https://github.com/golang/go",
			Description: stringPtr("The Go programming language repository on GitHub"),
			Tags:        []string{"development", "programming", "golang", "opensource"},
		},
		{
			Title:       "Stack Overflow - Programming Q&A",
			URL:         "https://stackoverflow.com",
			Description: stringPtr("The largest online community for programmers to learn and share knowledge"),
			Tags:        []string{"development", "qa", "community", "programming"},
		},
		{
			Title:       "MDN Web Docs",
			URL:         "https://developer.mozilla.org",
			Description: stringPtr("Resources for developers, by developers"),
			Tags:        []string{"web", "javascript", "html", "css", "documentation"},
		},
		{
			Title:       "Docker Hub",
			URL:         "https://hub.docker.com",
			Description: stringPtr("Container registry and development platform"),
			Tags:        []string{"docker", "containers", "devops", "deployment"},
		},
		{
			Title:       "Kubernetes Documentation",
			URL:         "https://kubernetes.io/docs/",
			Description: stringPtr("Production-grade container orchestration"),
			Tags:        []string{"kubernetes", "devops", "containers", "orchestration"},
		},
		{
			Title:       "Hacker News",
			URL:         "https://news.ycombinator.com",
			Description: stringPtr("Social news website focusing on computer science and entrepreneurship"),
			Tags:        []string{"news", "tech", "startup", "community"},
		},
		{
			Title:       "TypeScript Handbook",
			URL:         "https://www.typescriptlang.org/docs/",
			Description: stringPtr("TypeScript language documentation and guides"),
			Tags:        []string{"typescript", "javascript", "programming", "documentation"},
		},
		{
			Title:       "SQLite Documentation",
			URL:         "https://www.sqlite.org/docs.html",
			Description: stringPtr("SQLite database documentation"),
			Tags:        []string{"sqlite", "database", "sql", "documentation"},
		},
		{
			Title:       "Vite - Frontend Tooling",
			URL:         "https://vitejs.dev",
			Description: stringPtr("Next generation frontend tooling"),
			Tags:        []string{"vite", "frontend", "build-tools", "javascript"},
		},
		{
			Title:       "Tailwind CSS",
			URL:         "https://tailwindcss.com",
			Description: stringPtr("A utility-first CSS framework"),
			Tags:        []string{"css", "tailwind", "frontend", "ui", "design"},
		},
	}

	created := 0
	for _, bookmark := range demoBookmarks {
		// Check if bookmark already exists
		if existing, _ := bookmarkRepo.GetByURL(bookmark.URL); existing != nil {
			fmt.Printf("⏭️  Skipping existing bookmark: %s\n", bookmark.Title)
			continue
		}

		// Create bookmark
		if _, err := bookmarkRepo.Create(&bookmark); err != nil {
			fmt.Printf("❌ Failed to create bookmark %s: %v\n", bookmark.Title, err)
			continue
		}

		fmt.Printf("✅ Created bookmark: %s\n", bookmark.Title)
		created++
	}

	fmt.Printf("\n🎉 Demo data seeding complete! Created %d new bookmarks.\n", created)
	fmt.Printf("🚀 Start the server with: ./torimemo\n")
	fmt.Printf("🌐 Open: http://localhost:8080\n")
}