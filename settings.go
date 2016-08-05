package main

var cmdSettings = &Command{
	Run:   runSettings,
	Usage: "settings [index]",
	Short: "prints index settings",
	Long: `
Prints index settings.

Example:

  $ es settings
  $ es settings twitter
`,
	ApiUrl: "http://www.elasticsearch.org/guide/reference/api/admin-indices-get-settings.html",
}

func runSettings(cmd *Command, args []string) {
	index := ""
	if len(args) >= 1 {
		index = args[0]
	}

	logJson(esClient().IndexGetSettings(index).Do())
}
