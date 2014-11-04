package main

import (
	"errors"
	"flag"
	"fmt"
	notify "github.com/mqu/go-notify"
	"os"
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

const (
	appName      = "Tea Time(r)"
	rdyMsg       = "Your tea is ready! Enjoy :)"
	notifTimeout = 3000
)

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
	scale := 10.0

	// calculate what percent of the total time has passed
	perc := (remainingTime.Seconds() / totalTime.Seconds()) * 100
	// and scale it down
	percScaled := int(scale - (perc / scale))

	// generate the progress bar
	progressBar := strings.Repeat("#", percScaled) + strings.Repeat("-", int(scale)-percScaled)

	// and the entire progress line
	progress := fmt.Sprintf("Progress: [%s] (%3d%%) | %3.0f/%3.0f seconds remaining", progressBar, int(100-perc), remainingTime.Seconds(), totalTime.Seconds())

	fmt.Printf("\r%s", progress)
}

func notifyReady() {
	notify.Init(appName)
	notif := notify.NotificationNew(appName, rdyMsg, "")
	notif.SetTimeout(notifTimeout)
	if notif == nil {
		fmt.Println("Failed to create notification")
		return
	}

	if err := notif.Show(); err != nil && err.GError != nil {
		fmt.Printf("Error showing notification! Error was: %#v\n", err)
	}

	if err := notif.Close(); err != nil && err.GError != nil {
		fmt.Printf("Error closing notification channel! Error was: %#v\n", err)
	}
}

func main() {
	parseFlags()

	if listTeas {
		printLogo()
		printTeas()
		os.Exit(0)
	}

	if durationArg == 0 && tTypeArg == "" {
		fmt.Errorf("No tea type/custom duration supplied!\n")
		flag.Usage()
		os.Exit(1)
	}

	var duration time.Duration
	if durationArg != 0 {
		duration = durationArg
	} else {
		tType, err := parseTeaType(tTypeArg)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		duration = tType.steepTime
		fmt.Println(tType)
	}

	// add one second more for the progress bar to reach 100%
	timer := time.NewTimer(duration + time.Second)
	remainingTime := duration

	printLogo()

loop:
	for {
		select {
		case <-timer.C:
			fmt.Println()
			break loop
		default:
			printProgress(remainingTime, duration)
			time.Sleep(time.Second)
			remainingTime -= time.Second
		}
	}

	notifyReady()
	fmt.Println(rdyMsg)
}
