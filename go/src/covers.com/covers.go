package main

import (
	"os"
	"bufio"
	"fmt"
	"time"
	"strings"
	"strconv"

    "github.com/sclevine/agouti"
    "go.uber.org/zap"
    "gopkg.in/natefinch/lumberjack.v2"
    "go.uber.org/zap/zapcore"
)

var lumlog = &lumberjack.Logger{
    Filename:   "/tmp/covers.log",
    MaxSize:    10, // megabytes
    MaxBackups: 3,  // number of log files
    MaxAge:     3,  // days
}

func lumberjackZapHook(e zapcore.Entry) error {
    lumlog.Write([]byte(fmt.Sprintf("%+v", e)))
    return nil
}

func saveSource(p *agouti.Page, logger *zap.Logger, path string) {
	f, err := os.Create(path)
	if err != nil {
		logger.Fatal("Failed to create file")
	}

	defer f.Close()

	ps, err := p.HTML()

	w := bufio.NewWriter(f)
	n, err := w.WriteString(ps)
	if err != nil {
		logger.Fatal("Failed to write string")
	}
	fmt.Printf("wrote %d bytesn", n)

	w.Flush()
}

func navigatePatiently(url string, p *agouti.Page, logger *zap.Logger) {
	if err := p.Navigate(url); err != nil {
		logger.Fatal("Failed to navigate")
	}
	time.Sleep(2 * time.Second)
}

func setInputElementById(p *agouti.Page, logger *zap.Logger, id string, option string) {
	element := p.FindByID(id)
	element.Select(option)
	value, err := element.Attribute("value")

	if(err != nil) {
		logger.Error("Unable to set element id " + id + " to value: " + value)
	} else {
		// logger.Info("Element ID " + id + " set to value: " + value)
		path := "/tmp/" + time.Now().Format("2006-01-02_150405") + ".html"
		saveSource(p, logger, path)
	}
}

func getElementTextByClass(p *agouti.Page, logger *zap.Logger, class string) string {
	element := p.FindByClass(class)
	elementText, _ := element.Text()

	if len(elementText) > 0 {
		// logger.Info("Text: " + elementText)
		return elementText
	} else {
		// logger.Info("Unable to get element text by class: " + class)
		path := "/tmp/" + time.Now().Format("2006-01-02_150405") + ".html"
		saveSource(p, logger, path)
		return ""
	}
}

func getLastUpdated(p *agouti.Page, logger *zap.Logger) time.Time {
	text := getElementTextByClass(p, logger, "covers-CoversOdds-lastUpdated")
	if(text != "") {
		parts := strings.Split(text, "\n")

		//get current year using current eastern time (covers doesn't show the year so we need to get it ourselves)
		loc, _ := time.LoadLocation("America/New_York")
		now := time.Now().In(loc)
		currentYear := now.Year()

		//Parse the timestamp shown with the current year prepended
		form := "2006 Jan 2, 3:04 PM ET"
	    lastUpdated, _ := time.ParseInLocation(form, strconv.Itoa(currentYear) + " " + parts[1], loc)
	    // logger.Info("Last updated: " + lastUpdated.String())

	    //Convert the last updated time to UTC
	    loc, _ = time.LoadLocation("UTC")
    	lastUpdated = lastUpdated.In(loc)
	    // logger.Info("Last updated: " + lastUpdated.String())
	    return lastUpdated
	} else {
		return time.Now()
	}
}

func main() {
	logger, _ := zap.NewProduction(zap.Hooks(lumberjackZapHook))
	defer logger.Sync()

	driver := agouti.ChromeDriver()
	// headless browser
	// driver := agouti.ChromeDriver(
	//   agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu", "--no-sandbox"}),
	// )

	if err := driver.Start(); err != nil {
		logger.Fatal("Failed to start driver")
	}

	page, err := driver.NewPage()
	if err != nil {
		logger.Fatal("Failed to open page")
	}
	page.Session().SetImplicitWait(10)

	//navigate to covers.com mlb odds
	url := "https://www.covers.com/Sports/MLB/Odds/US/MONEYLINE/competition/Online/ML"
	navigatePatiently(url, page, logger)

	//set book type (Vegas/Offshore)
	setInputElementById(page, logger, "bookType", "Vegas")

	//set odds scope (competition/Futures)
	setInputElementById(page, logger, "oddScope", "competition")

	//set odd type (MONEYLINE/SPREAD/TOTAL)
	setInputElementById(page, logger, "oddType", "MONEYLINE")

	//set odd converter (ML (american)/EU (decimal)/UK (fractional))
	setInputElementById(page, logger, "oddConverter", "ML")

	//get last updated
	lastUpdated := getLastUpdated(page, logger)
	fmt.Println("last updated: " + lastUpdated.String())
	
	if err := page.Destroy(); err != nil {
		logger.Fatal("Failed to close pages")
	}

	if err := driver.Stop(); err != nil {
		logger.Fatal("Failed to stop WebDriver")
	}
}