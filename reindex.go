package main

import (
	"fmt"
	"log"
	_ "net/http/httputil"
	"os"

	"gopkg.in/olivere/elastic.v3"
)

var cmdReindex = &Command{
	Run:   runReindex,
	Usage: "reindex [-v] [-bulk=<n>] [-shards=<n>] [-replicas=<n>] <source> <target>",
	Short: "reindex one index to another index",
	Long: `
The reindex command takes the documents from the source index
and bulk imports them to the specified target index.

This is quite handy if you want to change the settings of an
index and won't lose any data.

You can also copy indices from one cluster to another by using the -source
and -target options.

Specifying -shards and/or -replicas is also possible; however, the target
index must not exist. Otherwise, -shards/-replicas will not be changed in
the target index.

Use the -v flag to print progress.

Example:

  $ es reindex -v twitter twitter-snapshot
  $ es reindex -v -source=http://cluster1:9200 -target=http://cluster2:9200 twitter twitter-snapshot
`,
	ApiUrl: "http://www.elasticsearch.org/guide/reference/api/bulk.html",
}

var (
	sourceURL, targetURL string
	bulkSize             int
	shards, replicas     int
)

func init() {
	cmdReindex.Flag.BoolVar(&verbose, "v", false, "verbose")
	cmdReindex.Flag.StringVar(&sourceURL, "source", "", "URL of source cluster")
	cmdReindex.Flag.StringVar(&targetURL, "target", "", "URL of target cluster")
	cmdReindex.Flag.IntVar(&bulkSize, "bulk", 1000, "bulk size")
	cmdReindex.Flag.IntVar(&shards, "shards", -1, "number of shards for target index")
	cmdReindex.Flag.IntVar(&replicas, "replicas", -1, "number of replicas for target index")
}

func runReindex(cmd *Command, args []string) {
	if len(args) < 2 {
		cmd.printUsage()
		os.Exit(1)
	}

	sourceIndex := args[0]
	targetIndex := args[1]

	if sourceURL == "" {
		sourceURL = esUrl
	}
	if targetURL == "" {
		targetURL = esUrl
	}

	// Get a client
	sourceClient, err := elastic.NewClient(elastic.SetURL(sourceURL))
	if err != nil {
		log.Fatalf("%v", err)
	}
	targetClient := sourceClient
	if sourceURL != targetURL {
		targetClient, err = elastic.NewClient(elastic.SetURL(targetURL))
		if err != nil {
			log.Fatalf("%v", err)
		}
	}

	// Check if source index exists. Stop if it doesn't.
	exists, err := sourceClient.IndexExists(sourceIndex).Do()
	if err != nil {
		log.Fatalf("%v", err)
	}
	if !exists {
		os.Exit(0)
	}

	// Check if the target index exists. If it doesn't, create it with
	// the number of shards/replicas specified.
	exists, err = targetClient.IndexExists(targetIndex).Do()
	if err != nil {
		log.Fatalf("%v", err)
	}
	if !exists {
		settings := make(map[string]interface{})
		if shards > 0 || replicas >= 0 {
			ixs := make(map[string]interface{})
			if shards > 0 {
				ixs["number_of_shards"] = shards
			}
			if replicas >= 0 {
				ixs["number_of_replicas"] = replicas
			}
			settings["index"] = ixs
		}
		_, err = targetClient.CreateIndex(targetIndex).BodyJson(settings).Do()
		if err != nil {
			log.Fatalf("%v", err)
		}
	} else if shards > 0 || replicas >= 0 {
		fmt.Fprint(os.Stderr, "Shards and/or replicas will not be changed on an existing index\n")
	}

	// Progress callback
	progress := func(current, total int64) {
		if verbose {
			var percent int64
			if total > 0 {
				percent = 100 * current / total
			}
			fmt.Fprintf(os.Stderr, "Reindexing %9d of %9d (%3d%%)\r", current, total, percent)
		}
	}

	// Use the Elastic Reindexer
	ix := elastic.NewReindexer(sourceClient, sourceIndex, elastic.CopyToTargetIndex(targetIndex))
	ix = ix.TargetClient(targetClient)
	if bulkSize > 0 {
		ix = ix.BulkSize(bulkSize)
	}
	ix = ix.Scroll("5m")
	ix = ix.Progress(progress)
	ix = ix.StatsOnly(true)
	res, err := ix.Do()
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Fprintf(os.Stderr, "%d successful and %d failed request(s)\n", res.Success, res.Failed)
}
