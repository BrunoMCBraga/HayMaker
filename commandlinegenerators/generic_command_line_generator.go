package commandlinegenerators

import (
	"flag"
	"strings"

	"github.com/haymaker/globalstringsproviders"
)

var option *string
var configFile *string
var kubeconfigFile *string

func PrepareCommandLineProcessing() {

	optionHelp := globalstringsproviders.GetOptionsMenu()

	option = flag.String("command", "", strings.TrimLeft(optionHelp, "\n"))
	configFile = flag.String("config", "", "Configuration file containing specs about: aws, networks, kubernetes cluster, node groups, containers.")
	kubeconfigFile = flag.String("kubeconfig", "", "Path used to save the Kubeconfig file. Used with gks option. If not provided, the default ~/.kube/config will be used.")
}

func ParseCommandLine() {
	flag.Parse()
}

func GetParametersDict() map[string]*string {

	parameters := make(map[string]*string, 0)
	parameters["option"] = option
	parameters["config_file"] = configFile
	parameters["kubeconfig_file"] = kubeconfigFile

	return parameters
}
