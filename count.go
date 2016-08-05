package main

import (
	"os"
)

var cmdCount = &Command{
	Run:   runCount,
	Usage: "count <index>",
	Short: "count documents in indices",
	Long: `
Returns the number of documents in one or more indices.

Example:

  $ es count catalog-1
  $ es count "catalog-*"
`,
	ApiUrl: "http://www.elasticsearch.org/guide/en/elasticsearch/reference/current/search-count.html",
}

func runCount(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.printUsage()
		os.Exit(1)
	}

	logJson(esClient().Count(args[0]).Do())
}
