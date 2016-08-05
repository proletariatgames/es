package main

var cmdRefresh = &Command{
	Run:   runRefresh,
	Usage: "refresh [index]",
	Short: "refreshes indices",
	Long: `
Refresh allows to explicitly refresh one or more index, 
making all operations performed since the last refresh 
available for search.

Example:

  $ es refresh
  $ es refresh twitter
`,
	ApiUrl: "http://www.elasticsearch.org/guide/reference/api/admin-indices-refresh.html",
}

func runRefresh(cmd *Command, args []string) {
	index := ""
	if len(args) >= 1 {
		index = args[0]
	}

	logJson(esClient().Refresh(index).Do())
}
