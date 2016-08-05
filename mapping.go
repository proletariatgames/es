package main

import (
	"os"
)

var cmdMapping = &Command{
	Run:   runMapping,
	Usage: "mapping <index>",
	Short: "list mapping",
	Long: `
Lists the mapping of an index.

Example:

  $ es mapping twitter
`,
	ApiUrl: "http://www.elasticsearch.org/guide/reference/api/admin-indices-get-mapping.html",
}

func runMapping(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.printUsage()
		os.Exit(1)
	}

	logJson(esClient().GetMapping().Index(args[0]).Do())
}
