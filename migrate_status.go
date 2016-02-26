package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/olivere/elastic"
)

var cmdMigrateStatus = &Command{
	Run:   runMigrateStatus,
	Usage: "migrate-status <directory> <environment>",
	Short: "migration status",
	Long: `
Queries elasticsearch for the status of the migrations in the given directory.

Example:

  $ es migrate-status migrations dev
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
		Env     string     `json:"env"`
	}
	migration struct {
		Version   int64
		AppliedAt *time.Time
		Source    string // path to .json script
	}
	migrationSorter []*migration
)

var migrationIndex string = ".es-migrate"
var migrationIndexJson string = `{
  "mappings": {
     "_default_": {
      "_all": {
        "norms": {
          "enabled": false
        }
      },
      "properties": {
        "migration": {
          "properties": {
            "when":    {"type": "date"},
            "version": {"type": "long"},
            "env":     {"type": "string", "index": "not_analyzed"}
		      }
		    }
		  }
    }
  }
}`

// helpers so we can use pkg sort
func (ms migrationSorter) Len() int           { return len(ms) }
func (ms migrationSorter) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms migrationSorter) Less(i, j int) bool { return ms[i].Version < ms[j].Version }

func migrationsInit(client *elastic.Client) error {
	exists, err := client.IndexExists(migrationIndex).Do()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = client.CreateIndex(migrationIndex).BodyString(migrationIndexJson).Do()
	return err
}

func migrationUpdate(record *migrationRecord, environment string, client *elastic.Client) error {
	_, err := client.Index().Index(migrationIndex).Type("migration").BodyJson(record).Do()
	return err
}

func migrationsLoad(directory, environment string, client *elastic.Client) ([]*migration, int, error) {
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

	pending := len(found)
	exists, err := client.IndexExists(migrationIndex).Do()
	if err != nil {
		return nil, 0, err
	}
	if exists {
		filter := elastic.NewTermFilter("env", environment)
		scroll := client.Scroll().Index(migrationIndex).Type("migration").Query(filter)
		for {
			result, err := scroll.Do()
			if err == elastic.EOS {
				break
			}
			if err != nil {
				return nil, 0, err
			}

			scroll = scroll.ScrollId(result.ScrollId)
			var record migrationRecord
			for _, item := range result.Each(reflect.TypeOf(record)) {
				if r, ok := item.(migrationRecord); ok {
					// TODO binary search?
					for _, migration := range found {
						if r.Version == migration.Version {
							migration.AppliedAt = r.When
							pending--
						}
					}
				}
			}
		}
	}

	return found, pending, nil
}

func runMigrateStatus(cmd *Command, args []string) {
	if len(args) < 2 {
		cmd.printUsage()
		os.Exit(1)
	}

	directory := args[0]
	environment := args[1]

	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(esUrl))
	if err != nil {
		log.Fatal(err)
	}

	migrations, _, err := migrationsLoad(directory, environment, client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("migrate: status for url '%v' environment '%v'\n", esUrl, environment)
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
