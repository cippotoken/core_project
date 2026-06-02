package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"wallet-generator-scanner/generator"
	"wallet-generator-scanner/githelper"
)

const (
	defaultEnvFile = ".env"
	defaultCount   = 100000
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Flags for the commands
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	countFlag := generateCmd.Int("count", defaultCount, "Number of private keys to generate")

	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	intervalFlag := runCmd.Duration("interval", 30*time.Second, "Interval between generation & push cycles")

	switch command {
	case "generate":
		_ = generateCmd.Parse(os.Args[2:])
		err := generator.GenerateWalletsToFile(defaultEnvFile, *countFlag)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "push":
		if len(os.Args) < 3 {
			fmt.Println("Error: Missing GitHub repository URL.")
			fmt.Println("Usage: go run main.go push <github-repo-url>")
			os.Exit(1)
		}
		repoURL := os.Args[2]
		cwd, _ := os.Getwd()

		err := githelper.SetupGitRepo(cwd, repoURL)
		if err != nil {
			fmt.Printf("Error setting up Git repository: %v\n", err)
			os.Exit(1)
		}

		err = githelper.CommitAndPush(cwd)
		if err != nil {
			fmt.Printf("Error committing and pushing: %v\n", err)
			os.Exit(1)
		}

	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Error: Missing GitHub repository URL.")
			fmt.Println("Usage: go run main.go run <github-repo-url> [-interval=30s]")
			os.Exit(1)
		}
		repoURL := os.Args[2]
		_ = runCmd.Parse(os.Args[3:])

		cwd, _ := os.Getwd()
		fmt.Printf("Starting Wallet Generator & GitHub Pusher daemon...\n")
		fmt.Printf("Target GitHub Repo: %s\n", repoURL)
		fmt.Printf("Cycle Interval: %v\n", *intervalFlag)

		// Setup repo initially
		err := githelper.SetupGitRepo(cwd, repoURL)
		if err != nil {
			fmt.Printf("Initial Git setup failed: %v\n", err)
			os.Exit(1)
		}

		// Run immediate first cycle
		runCycle(cwd, *intervalFlag)

		// Start ticker
		ticker := time.NewTicker(*intervalFlag)
		fmt.Printf("\n[Daemon] Next cycle scheduled in %v...\n", *intervalFlag)

		for range ticker.C {
			runCycle(cwd, *intervalFlag)
			fmt.Printf("\n[Daemon] Next cycle scheduled in %v...\n", *intervalFlag)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runCycle(cwd string, interval time.Duration) {
	fmt.Printf("\n--- START CYCLE: %s ---\n", time.Now().Format("2006-01-02 15:04:05"))
	
	// 1. Generate 100k new wallets
	err := generator.GenerateWalletsToFile(defaultEnvFile, defaultCount)
	if err != nil {
		fmt.Printf("[CYCLE ERROR] Generating wallets failed: %v\n", err)
		return
	}

	// 2. Commit and Push
	err = githelper.CommitAndPush(cwd)
	if err != nil {
		fmt.Printf("[CYCLE ERROR] Git push failed: %v\n", err)
		return
	}

	fmt.Printf("--- END CYCLE SUCCESSFUL ---\n")
}

func printUsage() {
	fmt.Println("Usage: go run main.go <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  generate [-count=100000]                Generate keys to .env without pushing")
	fmt.Println("  push <github-repo-url>                 Init git, commit, and push current .env")
	fmt.Println("  run <github-repo-url> [-interval=30s]   Start daemon: Generate 100k keys and push to GitHub every 30s")
}
