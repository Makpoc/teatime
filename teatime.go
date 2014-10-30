package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type tea struct {
	id        int
	ttype     string
	name      string
	steepTime time.Duration
	temp      int
}

func (t tea) String() string {
	return fmt.Sprintf(
		"Id:\t\t%d\nName:\t\t%s\nType:\t\t%s\nSteep Time:\t%.0f minutes\nTempreture:\t%d\u00B0",
		t.id,
		t.name,
		t.ttype,
		t.steepTime.Minutes(),
		t.temp)
}

var allTypes = []tea{
	tea{id: 0, ttype: "White", name: "White Dragon", steepTime: time.Minute * 2, temp: 70},
	tea{id: 1, ttype: "Green", name: "Temple of Heaven", steepTime: time.Minute * 3, temp: 80},
	tea{id: 2, ttype: "Green", name: "Green Dragon", steepTime: time.Minute * 3, temp: 80},
	tea{id: 3, ttype: "Black", name: "Lapsang Souchong", steepTime: time.Minute * 3, temp: 100},
	tea{id: 4, ttype: "Black", name: "Greenfield Magic Yunnan", steepTime: time.Minute * 7, temp: 100},
}

var durationArg time.Duration
var tTypeArg string
var listTeas bool

func printLogo() {
	fmt.Println(`
      Tea Time(r)
         ____ 
      ,|'----'|
     ((|      |
      \|      |
       |      |
       '------'
     ^^^^^^^^^^^^`)
}

func printTeas() {
	for i, tType := range allTypes {
		if i != 0 {
			fmt.Println("------")
		}
		fmt.Println(tType)
	}
}

func parseFlags() {

	flag.StringVar(&tTypeArg, "type", "", "Type of Tea (either the name or the ID. See -list)")
	flag.DurationVar(&durationArg, "duration", 0, "Tee Timer Duration. Warning: This argument has higher priority than -type!")
	flag.BoolVar(&listTeas, "list", false, "List all available tea types and exit")

	flag.Parse()
}

func parseTeaType(ttype string) (tea, error) {
	ttype = strings.ToLower(strings.TrimSpace(ttype))
	for _, t := range allTypes {
		if ttype == strings.ToLower(t.name) {
			return t, nil
		}
	}

	errMsg := fmt.Sprintf("Unknown tea: %s", ttype)

	// try to convert ttype to integer and check if an ID is provided instead
	tId, err := strconv.Atoi(ttype)
	if err != nil {
		return tea{}, errors.New(errMsg)
	}

	for _, t := range allTypes {
		if tId == t.id {
			return t, nil
		}
	}

	return tea{}, errors.New(errMsg)
}

func printProgress(remainingTime, totalTime time.Duration) {

	perc := int((remainingTime.Seconds() / totalTime.Seconds()) * 100)

	// progess is downscaled to 10 and is counting forward, and remaining time is backwards
	percScale := 10 - (perc / 10)
	progressBar := fmt.Sprintf("%s%s", strings.Repeat("#", percScale), strings.Repeat("-", 10-percScale))

	progress := fmt.Sprintf("Progress: [%s] (%%%3d) | %3.0f/%3.0f seconds remaining", progressBar, 100-perc, remainingTime.Seconds(), totalTime.Seconds())

	fmt.Printf("\r%s", progress)
	//fmt.Printf("\r%3.0f seconds remaining", remainingTime.Seconds())
}

func main() {
	parseFlags()

	if listTeas {
		printLogo()
		printTeas()
		return
	}

	if durationArg == 0 && tTypeArg == "" {
		fmt.Errorf("No tea type/custom duration supplied!\n")
		flag.Usage()
		return
	}

	var duration time.Duration
	if durationArg != 0 {
		duration = durationArg
	} else {
		tType, err := parseTeaType(tTypeArg)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		duration = tType.steepTime
		fmt.Println(tType)
	}

	timer := time.NewTimer(duration)
	doneChan := make(chan bool)
	remainingTime := duration

	printLogo()

	go func() {
		for {
			select {
			case <-doneChan:
				// print the last percent and return
				printProgress(remainingTime, duration)
				fmt.Println()
				return
			default:
				printProgress(remainingTime, duration)
				time.Sleep(time.Second * 1)
				remainingTime -= time.Second * 1
			}
		}
	}()

	// block
	<-timer.C
	doneChan <- true

	rdyMsg := fmt.Sprintf("Tee is ready!")
	fmt.Println(rdyMsg)
}
