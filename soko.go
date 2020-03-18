package main

import (
	"fmt"
	"os"
	"soko/pkg/app"
	"soko/pkg/portage"
	"time"
)

func printHelp(){
	fmt.Println("Please specific one of the following options:")
	fmt.Println("  soko update  -- update the database")
	fmt.Println("  soko serve   -- serve the application")
}

func isCommand(command string) bool{
	return len(os.Args) > 1 && os.Args[1] == command
}

func main() {

	time.Sleep(5 * time.Second)

	if(isCommand("serve")) {
		app.Serve()
	}else if(isCommand("update")){
		portage.Update()
	}else{
		printHelp()
	}

}
