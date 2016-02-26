package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

var cmdCreateMigration = &Command{
	Run:   runCreateMigration,
	Usage: "create-migration <directory> <migration>",
	Short: "create migration",
	Long: `
Creates a new Migration.

Example:

  $ es create-migration migrations initial
`,
}

var migrationTemplate = template.Must(template.New("es.migration").Parse(`
{
  "up": {
    "endpoint": "<insert the endpoint here, index is passed in as a parameter>",
    "method": "POST|GET|PUT|DELETE",
    "payload": {},
  }
}
`))

func runCreateMigration(cmd *Command, args []string) {
	if len(args) < 2 {
		cmd.printUsage()
		os.Exit(1)
	}

	directory := args[0]
	migration := args[1]

	if err := os.MkdirAll(directory, 0777); err != nil {
		log.Fatal(err)
	}

	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%v_%v.json", timestamp, migration)
	fpath := filepath.Join(directory, filename)

	f, e := os.Create(fpath)
	if e != nil {
		log.Fatal(e)
	}
	defer f.Close()

	e = migrationTemplate.Execute(f, nil)
	if e != nil {
		log.Fatal(e)
	}

	a, e := filepath.Abs(fpath)
	if e != nil {
		log.Fatal(e)
	}

	fmt.Println("created", a)
}
