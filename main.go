package main

import (
	"errors"
	"fmt"

	"github.com/haymaker/commandlineparsers"
	"github.com/haymaker/commandlineprocessors"
	"github.com/haymaker/globalstringsproviders"
)

func main() {

	commandlineparsers.PrepareCommandLineProcessing()

	fmt.Println(globalstringsproviders.GetMenuPictureString())
	commandlineparsers.ParseCommandLine()
	parameters := commandlineparsers.GetParametersDict()
	processCommandLineProcessorError := commandlineprocessors.ProcessCommandLine(parameters)
	if processCommandLineProcessorError != nil {
		fmt.Println(errors.New("HayMaker->main->commandlineprocessors.ProcessCommandLine:" + processCommandLineProcessorError.Error()))
	}

}
