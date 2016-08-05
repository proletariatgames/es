package main

var cmdFlush = &Command{
	Run:   runFlush,
	Usage: "flush [index]",
	Short: "flushes indices",
	Long: `
Flushes data to the index storage, frees memory and clears the
internal transaction log. You often need to flush before updating
ElasticSearch to a new version.

Example:

  $ es flush
  $ es flush twitter
`,
	ApiUrl: "http://www.elasticsearch.org/guide/reference/api/admin-indices-flush.html",
}

func runFlush(cmd *Command, args []string) {
	index := ""
	if len(args) >= 1 {
		index = args[0]
	}

	f := esClient().Flush()
	if len(index) > 0 {
		f = f.Index(index)
	}

	logJson(f.Do())
}
