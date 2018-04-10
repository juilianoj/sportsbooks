package main

import (
	//"os"
	//"bufio"
	"fmt"
	"strings"
	"strconv"
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
	attempts := 0
	for attempts < 3 {
		attempts++
		url := "https://sports.bovada.lv/baseball/mlb/game-lines-market-group"
		navigatePatiently(url, p, logger)
		
		//get event count
		eventCntElem := p.First("#spaNavigationComponents_content_center > div > section > div > div > h2")
		eventCntTxt, _ := eventCntElem.Text()
		eventCnt := strings.Fields(eventCntTxt)

		//print to console, or retry
		if len(eventCnt) > 0 {
			logger.Info("Event count: " + eventCnt[0])
			break
		} else {
			logger.Info("No events found, attempt: " + strconv.Itoa(attempts))
		}
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

func main() {
	logger, _ := zap.NewProduction(zap.Hooks(lumberjackZapHook))
	defer logger.Sync()

	driver := agouti.ChromeDriver()
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

	// //sample code to write page source to a file
	// f, err := os.Create("bovada.html")
	// if err != nil {
	// 	log.Fatal("Failed to create file:", err)
	// }

	// defer f.Close()

	// ps, err := page.HTML()

	// w := bufio.NewWriter(f)
	// n, err := w.WriteString(ps)
	// if err != nil {
	// 	log.Fatal("Failed to write string:", err)
	// }
	// fmt.Printf("wrote %d bytesn", n)

	// w.Flush()