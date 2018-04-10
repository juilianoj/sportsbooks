package main

import (
	"os"
	"bufio"
	"fmt"
	"strings"
	"time"

    "github.com/sclevine/agouti"
    "go.uber.org/zap"
    "gopkg.in/natefinch/lumberjack.v2"
    "go.uber.org/zap/zapcore"
)

func navigatePatiently(url string, p *agouti.Page, logger *zap.Logger) {
	if err := p.Navigate(url); err != nil {
		logger.Fatal("Failed to navigate")
	}
	time.Sleep(2 * time.Second)
}

func getBovadaMLBEventCount(p *agouti.Page, logger *zap.Logger) {
	url := "https://sports.bovada.lv/baseball/mlb/game-lines-market-group"
	navigatePatiently(url, p, logger)
	
	//get event count
	eventCntElem := p.First("#spaNavigationComponents_content_center > div > section > div > div > h2")
	eventCntTxt, _ := eventCntElem.Text()
	eventCnt := strings.Fields(eventCntTxt)

	if len(eventCnt) > 0 {
		logger.Info("Event count: " + eventCnt[0])
	} else {
		logger.Info("No events found, saving page source.")
		path := "/tmp/" + time.Now().Format("2006-01-02_150405") + ".html"
		saveSource(p, logger, path)
	}
}

var lumlog = &lumberjack.Logger{
    Filename:   "/tmp/bovada.log",
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

	getBovadaMLBEventCount(page, logger)
	
	if err := page.Destroy(); err != nil {
		logger.Fatal("Failed to close pages")
	}

	if err := driver.Stop(); err != nil {
		logger.Fatal("Failed to stop WebDriver")
	}
}