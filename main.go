package main

import (
	"errors"
	"fmt"

	"github.com/BrunoMCBraga/HayMaker/commandlinegenerators"
	"github.com/BrunoMCBraga/HayMaker/commandlineprocessors"
	"github.com/BrunoMCBraga/HayMaker/globalstringsproviders"
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
