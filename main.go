package main

import (
	"errors"
	"fmt"

	"github.com/haymaker/commandlinegenerators"
	"github.com/haymaker/commandlineprocessors"
	"github.com/haymaker/globalstringsproviders"
)

func main() {

	commandlinegenerators.PrepareCommandLineProcessing()

	fmt.Println(globalstringsproviders.GetMenuPictureString())
	commandlinegenerators.ParseCommandLine()
	parameters := commandlinegenerators.GetParametersDict()
	processCommandLineProcessorError := commandlineprocessors.ProcessCommandLine(parameters)
	if processCommandLineProcessorError != nil {
		fmt.Println(errors.New("HayMaker->main->commandlineprocessors.ProcessCommandLine:" + processCommandLineProcessorError.Error()))
	}

}

