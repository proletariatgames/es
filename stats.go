package main

var cmdStats = &Command{
	Run:   runStats,
	Usage: "stats [index]",
	Short: "prints index statistics",
	Long: `
Lists detailed information such as the number of documents
in an index etc.

Example:

  $ es stats
  $ es stats twitter
`,
	ApiUrl: "http://www.elasticsearch.org/guide/reference/api/admin-indices-stats.html",
}

func runStats(cmd *Command, args []string) {
	index := ""
	if len(args) >= 1 {
		index = args[0]
	}

	logJson(esClient().IndexStats(index).Do())
}
