package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
)

var cmdMigrateStatus = &Command{
	Run:   runMigrateStatus,
	Usage: "migrate-status <directory>",
	Short: "migration status",
	Long: `
Queries elasticsearch for the status of the migrations in the given directory.

Example:

  $ es migrate-status migrations
`,
}

type (
	migrationCommand struct {
		Endpoint string          `json:"endpoint"`
		Method   string          `json:"method"`
		Payload  json.RawMessage `json:"payload"`
	}
	migrationFile struct {
		Up *migrationCommand `json:"up"`
	}
	migrationRecord struct {
		When    *time.Time `json:"when"`
		Version int64      `json:"version"`
	}
	migration struct {
		Version   int64
		AppliedAt *time.Time
		Source    string // path to .json script
	}
	migrationSorter []*migration
)

// helpers so we can use pkg sort
func (ms migrationSorter) Len() int           { return len(ms) }
func (ms migrationSorter) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms migrationSorter) Less(i, j int) bool { return ms[i].Version < ms[j].Version }

func migrationsLoad(directory string) ([]*migration, error) {
	var found []*migration
	patt := regexp.MustCompile(`^(\d{14})_[a-z0-9_]+\.json$`)
	filepath.Walk(directory, func(name string, info os.FileInfo, err error) error {
		if patt.MatchString(info.Name()) {
			matches := patt.FindStringSubmatch(info.Name())
			version, _ := strconv.ParseInt(matches[1], 10, 64)
			found = append(found, &migration{version, nil, name})

		}
		return nil
	})

	sort.Sort(migrationSorter(found))
	return found, nil
}

func runMigrateStatus(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.printUsage()
		os.Exit(1)
	}

	directory := args[0]
	migrations, err := migrationsLoad(directory)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("migrate: status for url '%v' directory '%v'\n", esUrl, directory)
	fmt.Println("    Applied At                  Migration")
	fmt.Println("    =======================================")
	for _, m := range migrations {
		var appliedAt string

		if m.AppliedAt != nil {
			appliedAt = m.AppliedAt.Format(time.ANSIC)
		} else {
			appliedAt = "Pending"
		}

		fmt.Printf("    %-24s -- %v\n", appliedAt, m.Source)
	}
}
