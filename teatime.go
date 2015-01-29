package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	notify "github.com/mqu/go-notify"
)

// steepTime is a time.Duration, which implements the Unmarshaler interface
type steepTime struct {
	time.Duration
}

func (t *steepTime) UnmarshalJSON(data []byte) error {
	var s string

	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	t.Duration = dur
	return nil
}

// tea contains information about the type of tea as well as some details about the preparation
type tea struct {
	ID        int       `json:"id"`
	Ttype     string    `json:"type"`
	Name      string    `json:"name"`
	SteepTime steepTime `json:"steepTime"`
	Temp      int       `json:"temp"`
}

func (t tea) String() string {
	steepTimeTotal := t.SteepTime.Seconds()
	steepTimeFmt := fmt.Sprintf("%.0f minutes, %.0f seconds", steepTimeTotal/60.0, float32(int(steepTimeTotal)%60))

	return fmt.Sprintf(
		"ID:\t\t%d\nName:\t\t%s\nType:\t\t%s\nSteep Time:\t%s\nTemperature:\t%d\u00B0",
		t.ID,
		t.Name,
		t.Ttype,
		steepTimeFmt,
		t.Temp)
}

var defaultTeas = []tea{
	{ID: 0, Ttype: "White", Name: "White Dragon", SteepTime: steepTime{Duration: time.Minute * 2}, Temp: 70},
	{ID: 1, Ttype: "Green", Name: "Temple of Heaven", SteepTime: steepTime{Duration: time.Minute * 2}, Temp: 80},
	{ID: 2, Ttype: "Green", Name: "Green Dragon", SteepTime: steepTime{Duration: time.Minute * 2}, Temp: 80},
	{ID: 3, Ttype: "Black", Name: "Lapsang Souchong", SteepTime: steepTime{Duration: time.Minute * 2}, Temp: 100},
	{ID: 4, Ttype: "Black", Name: "Greenfield Magic Yunnan", SteepTime: steepTime{Duration: time.Minute * 7}, Temp: 100},
}

var durationArg string
var teaArg string
var listTeas bool
var filePath string

const (
	appName      = "Tea Time(r)"
	rdyMsg       = "Your tea is ready! Enjoy :)"
	notifTimeout = 3000
)

// printLogo prints the application logo on the console
func printLogo() {
	fmt.Println(`
      Tea Time(r)
         ____    ,-^-,
      ,|'----'|  * L *
     ((|      |  '-.-'
      \|      |
       |      |
       '------'
     ^^^^^^^^^^^^`)
}

// PrintTeas prints information about each tea on the console
func printTeas(teas []tea) {
	for i, tType := range teas {
		if i != 0 {
			fmt.Println("------")
		}
		fmt.Println(tType)
	}
}

// loadTeas reads the json stream and tries to decode all available teas from it. If this fails - it returns the default set of teas and an error
func loadTeas(reader io.Reader) ([]tea, error) {
	var allTeas []tea
	if err := json.NewDecoder(reader).Decode(&allTeas); err != nil {
		return defaultTeas, fmt.Errorf("Failed to parse file. Using default list of teas! Error was: %s", err.Error())
	}

	return allTeas, nil

}

// parseFlags parses the command line flags
func parseFlags() {

	flag.StringVar(&durationArg, "duration", "", "\tTee timer duration. Can be Xs/m/h (overwrite -tea's default duration if given) or +-Xs/m/h (add to it)")
	flag.StringVar(&teaArg, "tea", "", "\t\tType of Tea to prepare (either the Name or the ID. See -list)")
	flag.BoolVar(&listTeas, "list", false, "\t\tList all available tea types and exit with brief information about each of them")
	flag.StringVar(&filePath, "file", "", "\t\tPath to json file, containing tea specifications")

	flag.Parse()
}

func getTotalDuration(selectedTea tea, customDuration string) (time.Duration, error) {
	var baseDuration time.Duration
	if selectedTea != (tea{}) {
		baseDuration = selectedTea.SteepTime.Duration
	}

	var calcFunc func(time.Duration, time.Duration) (time.Duration, error)

	if strings.HasPrefix(customDuration, "+") {
		customDuration = strings.TrimLeft(customDuration, "+")
		calcFunc = addDur
	} else if strings.HasPrefix(customDuration, "-") {
		customDuration = strings.TrimLeft(customDuration, "-")
		calcFunc = subDur
	}

	customDur, err := time.ParseDuration(customDuration)
	if err != nil {
		return baseDuration, err
	}

	if calcFunc != nil {
		return calcFunc(baseDuration, customDur)
	}

	return customDur, nil
}

func addDur(baseDur, customDur time.Duration) (time.Duration, error) {
	return baseDur + customDur, nil
}

func subDur(baseDur, customDur time.Duration) (time.Duration, error) {
	if baseDur <= customDur {
		return baseDur, errors.New("Total duration must be positive!")
	}

	return baseDur - customDur, nil
}

func getTeaByID(id int, teas []tea) (tea, error) {
	for _, t := range teas {
		if id == t.ID {
			return t, nil
		}
	}

	return tea{}, fmt.Errorf("Tea with ID %d not found!", id)
}

func getTeaByName(name string, teas []tea) (tea, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, t := range teas {
		if name == strings.ToLower(t.Name) {
			return t, nil
		}
	}

	return tea{}, fmt.Errorf("Tea with Name %s not found!", name)
}

// getTea tries to find a specific tea in the provided list of teas. If ttype is of type int - it searches by ID. Otherwise it tries to exactly match the name.
func getTea(ttype string, teas []tea) (t tea, err error) {

	// Test if ttype contains a Name or ID
	if tID, err := strconv.Atoi(ttype); err == nil {
		return getTeaByID(tID, teas)
	}

	return getTeaByName(ttype, teas)
}

// printProgress displays a progress bar on the console.
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

// notifyReady shows a Desktop notification (if possible).
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

func getDurAndTea(allTeas []tea) (time.Duration, tea, error) {
	var emptyDuration time.Duration
	var emptyTea tea
	// verify that at least one of the arguments will set the duration. Otherwise there's no point in continuing
	if teaArg == "" && durationArg == "" {
		flag.Usage()
		os.Exit(1)
	}

	var teaType tea
	if teaArg != "" {
		// specific type of tea was requested - try to find it, update the duration and display details
		var err error
		teaType, err = getTea(teaArg, allTeas)
		if err != nil {
			return emptyDuration, emptyTea, err
		}
	}

	var duration time.Duration
	if durationArg != "" {
		// duration was provided on command line - overwrite the duration
		var err error
		duration, err = getTotalDuration(teaType, durationArg)
		if err != nil {
			return emptyDuration, emptyTea, err
		}
	} else {
		duration = teaType.SteepTime.Duration
	}

	return duration, teaType, nil
}

func main() {
	parseFlags()

	teas := defaultTeas

	if filePath != "" {
		// tea specs file was provided on command line - try to load it and parse it
		f, err := os.Open(filePath)
		defer f.Close()
		if err != nil {
			fmt.Printf("Failed to open file! Error was: %s\n", err.Error())
			os.Exit(1)
		}
		teas, err = loadTeas(f)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	if listTeas {
		printLogo()
		printTeas(teas)
		os.Exit(0)
	}

	duration, selectedTea, err := getDurAndTea(teas)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// if not empty
	if selectedTea != (tea{}) {
		fmt.Println(selectedTea)
	}
	printLogo()

	// add one second more for the progress bar to reach 100%
	timeout := time.After(duration + time.Second)
	remainingTime := duration

	// main progress loop
loop:
	for {
		select {
		case <-timeout:
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
