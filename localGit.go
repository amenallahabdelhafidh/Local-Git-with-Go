package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const folderListFile = "folders.txt"

func main() {
	var folder string
	var email string

	flag.StringVar(&folder, "add", "", "Add a new folder to scan for Git repositories")
	flag.StringVar(&email, "email", "", "Email to scan (required)")
	flag.Parse()

	if folder != "" {
		scan(folder)
		return
	}

	if email == "" {
		fmt.Println("Please provide an email using -email flag")
		return
	}

	stats(email)
}

// scan adds the folder path to folders.txt
func scan(folder string) {
	file, err := os.OpenFile(folderListFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening folder list file:", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(folder + "\n"); err != nil {
		fmt.Println("Error writing folder path:", err)
		return
	}

	fmt.Println("Folder added successfully:", folder)
}

// stats reads folders.txt and prints all commits in a table per folder
func stats(email string) {
	fmt.Println("Generating all-time commit stats for email:", email)

	file, err := os.Open(folderListFile)
	if err != nil {
		fmt.Println("Error opening folder list file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	folders := []string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			folders = append(folders, line)
		}
	}

	if len(folders) == 0 {
		fmt.Println("No folders to scan. Add folders using -add flag.")
		return
	}

	green := "\033[32m"
	reset := "\033[0m"

	for _, folder := range folders {
		if strings.HasPrefix(folder, "http") {
			fmt.Println("Skipping URL folder:", folder)
			continue
		}

		if _, err := os.Stat(folder); os.IsNotExist(err) {
			fmt.Println("Folder does not exist:", folder)
			continue
		}

		gitDir := filepath.Join(folder, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			fmt.Println("Not a Git repository:", folder)
			continue
		}

		cmd := exec.Command("git", "-C", folder, "log", "--all",
			"--author="+email,
			"--pretty=format:%ad|%s",
			"--date=short")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error running git in folder:", folder, err)
			continue
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
			fmt.Printf("Folder: %s -> No commits found for %s\n\n", folder, email)
			continue
		}

		fmt.Printf("Folder: %s\n", folder)
		fmt.Printf("| %-10s | %-50s |\n", "Date", "Commit Message")
		fmt.Println(strings.Repeat("-", 65))

		for _, line := range lines {
			parts := strings.SplitN(line, "|", 2)
			if len(parts) < 2 {
				continue
			}
			date := parts[0]
			message := parts[1]
			if len(message) > 50 {
				message = message[:47] + "..."
			}
			fmt.Printf("| %-10s | %s%-50s%s |\n", date, green, message, reset)
		}
		fmt.Println()
	}
}
