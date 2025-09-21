package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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
// stats reads folders.txt and prints a GitHub-style contributions heatmap
func stats(email string) {
	fmt.Println("Generating commit calendar for:", email)

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

	// Collect commits by date (YYYY-MM-DD)
	commitCount := make(map[string]int)
	for _, folder := range folders {
		gitDir := filepath.Join(folder, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			continue
		}

		cmd := exec.Command("git", "-C", folder, "log", "--all",
			"--pretty=format:%ad|%ae",
			"--date=short")

		output, err := cmd.Output()
		fmt.Println("DEBUG output from git log in", folder)
		fmt.Println(string(output))

		if err != nil {
			continue
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "|")
			if len(parts) != 2 {
				continue
			}
			date, author := parts[0], parts[1]
			if strings.EqualFold(author, email) { // case-insensitive match
				commitCount[date]++
			}
		}

	}

	// Start from 1 year ago (aligned to Sunday)
	today := time.Now()
	start := today.AddDate(0, 0, -365)
	for start.Weekday() != time.Sunday {
		start = start.AddDate(0, 0, -1)
	}

	// Print months header
	fmt.Print("     ")
	curMonth := ""
	for d := start; d.Before(today); d = d.AddDate(0, 0, 7) {
		month := d.Format("Jan")
		if month != curMonth {
			fmt.Printf("%-3s", month)
			curMonth = month
		} else {
			fmt.Print("   ")
		}
	}
	fmt.Println()

	// Print rows (Sun, Tue, Thu, Sat for compactness)
	weekdays := []time.Weekday{
		time.Sunday,
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
		time.Saturday,
	}

	for _, wd := range weekdays {
		fmt.Printf("%-3s ", wd.String()[:3])
		for d := start; d.Before(today); d = d.AddDate(0, 0, 7) {
			day := d
			for day.Weekday() != wd {
				day = day.AddDate(0, 0, 1)
			}
			dateStr := day.Format("2006-01-02")
			count := commitCount[dateStr]
			fmt.Print(colorForDate(count, dateStr))

		}
		fmt.Println()
	}

}

func colorForDate(count int, date string) string {
	var color string
	switch {
	case count == 0:
		color = "\033[48;5;232m"
	case count < 5:
		color = "\033[48;5;22m"
	case count < 10:
		color = "\033[48;5;28m"
	case count < 20:
		color = "\033[48;5;34m"
	default:
		color = "\033[48;5;40m"
	}

	if count == 0 {

		return fmt.Sprintf("%s  \033[0m", color)
	}

	day := date[len(date)-2:]
	return fmt.Sprintf("%s%s\033[0m", color, day)
}
