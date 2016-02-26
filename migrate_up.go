package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/olivere/elastic"
)

var cmdMigrateUp = &Command{
	Run:   runMigrateUp,
	Usage: "migrate-up <directory> <environment>",
	Short: "migration up",
	Long: `
Queries elasticsearch for the up of the migrations in the given directory.

Example:

  $ es migrate-up migrations up
`,
}

func migrationApply(migration *migration, when *time.Time, environment string, client *elastic.Client) error {
	tpl, err := template.ParseFiles(migration.Source)
	if err != nil {
		return err
	}

	var doc bytes.Buffer
	data := map[string]string{"Env": environment}
	if err := tpl.Execute(&doc, data); err != nil {
		return err
	}

	var file migrationFile
	if err := json.Unmarshal(doc.Bytes(), &file); err != nil {
		return err
	}

	body, err := file.Up.Payload.MarshalJSON()
	if err != nil {
		return err
	}

	if _, err := client.PerformRequest(file.Up.Method, file.Up.Endpoint, url.Values{}, string(body)); err != nil {
		return err
	}

	record := &migrationRecord{when, migration.Version, environment}
	if err := migrationUpdate(record, environment, client); err != nil {
		return err
	}

	return nil
}

func runMigrateUp(cmd *Command, args []string) {
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

	migrations, pending, err := migrationsLoad(directory, environment, client)
	if err != nil {
		log.Fatal(err)
	}

	if pending > 0 {
		if err := migrationsInit(client); err != nil {
			log.Fatal(err)
		}
	}

	when := time.Now()
	for _, migration := range migrations {
		if migration.AppliedAt != nil {
			continue
		}

		short := filepath.Base(migration.Source)
		if err := migrationApply(migration, &when, environment, client); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL  %v %v\n", short, err)
			os.Exit(1)
		}

		fmt.Println("OK   ", short)
	}
}
