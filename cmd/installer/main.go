package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/donkeysharp/donkeyvpn/internal/install"
	"github.com/labstack/gommon/log"
)

var tfvarsTemplate string
var tfvarsOutput string
var tfbackendTemplate string
var tfbackendOutput string

func init() {
	flag.StringVar(&tfvarsTemplate, "tfvars-template", "", "Location of the tfvars template file")
	flag.StringVar(&tfvarsOutput, "tfvars-output", "", "Location of the result tfvars file")
	flag.StringVar(&tfbackendTemplate, "tfbackend-template", "", "Location of the tfbackend template file")
	flag.StringVar(&tfbackendOutput, "tfbackend-output", "", "Location of the result tfbackend file")
}

func usage() {
	fmt.Println("Required parameters:")
	fmt.Println("\t-tfvars-template")
	fmt.Println("\t-tfvars-ouput")
	fmt.Println("\t-tfbackend-template")
	fmt.Println("\t-tfbackend-output")
}

func main() {
	log.SetLevel(log.OFF)
	flag.Parse()
	if tfvarsTemplate == "" || tfvarsOutput == "" ||
		tfbackendTemplate == "" || tfbackendOutput == "" {
		usage()
		os.Exit(1)
	}
	w := install.NewWizard(tfvarsTemplate, tfvarsOutput, tfbackendTemplate, tfbackendOutput)
	w.Start()
}
