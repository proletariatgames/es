package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/olivere/elastic.v3"
)

const (
	Version = "0.2.0"
)

var (
	verbose bool
	force   bool
	flush   bool
	refresh bool
)

type Command struct {
	Run  func(cmd *Command, args []string)
	Flag flag.FlagSet

	Usage string
	Short string
	Long  string

	ApiUrl string
}

func (c *Command) printUsage() {
	fmt.Printf("Usage: es %s\n\n", c.Usage)
	fmt.Println(strings.TrimSpace(c.Long))
}

func (c *Command) Name() string {
	name := c.Usage
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func getJsonFromStdin() interface{} {
	data := struct{}{}
	reader := bufio.NewReader(os.Stdin)
	stats, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("file.Stat()", err)
	}
	if stats.Size() > 0 {
		if err := json.NewDecoder(reader).Decode(&data); err != nil {
			log.Fatal("invalid json\n")
		}
	}
	return data
}

// Running es on the command line will print these commands in order.
var commands = []*Command{
	cmdSearch,
	cmdCount,

	cmdIndices,
	cmdCreateIndex,
	cmdDeleteIndex,
	cmdOpenIndex,
	cmdCloseIndex,
	cmdSettings,
	cmdStatus,
	cmdStats,
	cmdRefresh,
	cmdOptimize,
	cmdFlush,
	cmdFlushDisable,
	cmdFlushEnable,

	cmdAliases,
	cmdIndexAliases,

	cmdMapping,
	cmdPutMapping,

	cmdTemplates,
	cmdTemplate,
	cmdCreateTemplate,
	cmdDeleteTemplate,

	cmdBulk,
	cmdReindex,

	cmdClusterHealth,
	cmdClusterState,
	cmdClusterNodes,

	cmdRepos,
	cmdRepo,
	cmdCreateRepo,
	cmdDeleteRepo,

	cmdSnapshots,
	cmdSnapshot,
	cmdCreateSnapshot,
	cmdDeleteSnapshot,
	cmdRestoreSnapshot,
	cmdSnapshotStatus,

	cmdVersion,
	cmdApi,
	cmdHelp,
}

var (
	esUrl   string
	esSniff bool
)

func esClient() *elastic.Client {
	client, err := elastic.NewClient(elastic.SetSniff(esSniff), elastic.SetURL(esUrl))
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func main() {
	log.SetFlags(0)

	args := os.Args[1:]
	if len(args) < 1 {
		usage()
	}

	esUrl = elastic.DefaultURL
	esSniff = elastic.DefaultSnifferEnabled
	if s := os.Getenv("ES_URL"); s != "" {
		esUrl = strings.TrimRight(s, "/")
	}
	if s := os.Getenv("ES_SNIFF"); s != "" {
		esSniff = s == "true"
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] {
			cmd.Flag.Usage = usage
			cmd.Flag.Parse(args[1:])
			cmd.Run(cmd, cmd.Flag.Args())
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
	usage()
}
