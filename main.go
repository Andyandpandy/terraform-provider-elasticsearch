/*
Package main create Elasticsearch provider for Terraform

Read the doc to use it: https://github.com/disaster37/terraform-provider-elasticsearch/tree/7.x
*/
package main

import (
	"flag"
	"os"

	"github.com/disaster37/terraform-provider-elasticsearch/v8/es"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func init() {

	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&easy.Formatter{
		LogFormat: "[%lvl%] %msg%\n",
	})

}

func main() {

	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: es.Provider,
		Debug:        debugMode,
	}

	plugin.Serve(opts)

}
