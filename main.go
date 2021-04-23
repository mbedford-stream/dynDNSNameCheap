package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type jsonConfig struct {
	UpdateParams struct {
		Domain   string `json:"domain"`
		Host     string `json:"host"`
		Password string `json:"password"`
		Log      bool   `json:"log"`
		Debug    bool   `json:"debug"`
	} `json:"updateParams"`
}

type xmlResponse struct {
	XMLName  xml.Name `xml:"interface-response"`
	Text     string   `xml:",chardata"`
	Command  string   `xml:"Command"`
	Language string   `xml:"Language"`
	ErrCount string   `xml:"ErrCount"`
	Errors   struct {
		Text string `xml:",chardata"`
		Err1 string `xml:"Err1"`
	} `xml:"errors"`
	ResponseCount string `xml:"ResponseCount"`
	Responses     struct {
		Text     string `xml:",chardata"`
		Response struct {
			Text           string `xml:",chardata"`
			ResponseNumber string `xml:"ResponseNumber"`
			ResponseString string `xml:"ResponseString"`
		} `xml:"response"`
	} `xml:"responses"`
	Done  string `xml:"Done"`
	Debug string `xml:"debug"`
}

func FileExists(fileName string) bool {
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// FileIsADirectory - tests a file
func FileIsADirectory(file string) bool {
	if stat, err := os.Stat(file); err == nil && stat.IsDir() {
		// path is a directory
		return true
	}
	return false
}

// FileExistsAndIsNotADirectory - tests a file
func FileExistsAndIsNotADirectory(file string) bool {
	if FileExists(file) && !FileIsADirectory(file) {
		return true
	}
	return false
}

func myIP(ipURL string) (string, error) {

	var currentIP string

	getURL := ipURL
	res, err := http.Get(getURL)
	if err != nil {
		return currentIP, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return currentIP, err
	}
	currentIP = strings.Trim(string(body), "\n")

	return currentIP, nil
}

func currentDNS(hostCheck string) ([]string, error) {
	var dnsValue []string
	dnsValue, err := net.LookupHost(hostCheck)
	if err != nil {
		return dnsValue, err
	}
	return dnsValue, nil
}

func updateSend(hostUpdate string, domainUpdate string, passwordUpdate string, ipUpdate string) error {
	updateURLBase := "https://dynamicdns.park-your-domain.com/update?"
	updateVars := map[string]string{
		"host":     hostUpdate,
		"domain":   domainUpdate,
		"password": passwordUpdate,
		"ip":       ipUpdate}

	updateURL := updateURLBase
	var i int = 0
	for k, v := range updateVars {
		if i == len(updateVars)-1 {
			updateURL = fmt.Sprintf("%s%s=%s", updateURL, k, v)
		} else {
			updateURL = fmt.Sprintf("%s%s=%s&", updateURL, k, v)
		}
		i++
	}
	fmt.Println(updateURL)

	res, err := http.Get(updateURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var updateResult xmlResponse
	err = xml.NewDecoder(res.Body).Decode(&updateResult)
	if err != nil {
		return err
	}

	if updateResult.ErrCount != "0" {
		return errors.New(updateResult.Errors.Err1)
	}

	return nil
}

func readConfig(confFile string) (jsonConfig, error) {
	if !FileExistsAndIsNotADirectory(confFile) {
		log.Fatalf("Cannot find config file (%s)", confFile)
	}

	var configInfo jsonConfig

	rawConf, err := ioutil.ReadFile(confFile)
	if err != nil {
		return configInfo, err
	}

	confErr := json.Unmarshal(rawConf, &configInfo)
	if confErr != nil {
		return configInfo, err
	}

	return configInfo, nil
}

func writeLog(logFile string, passFail string, logData map[string]string) error {
	f, err := os.OpenFile(logFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	logger := log.New(f, passFail+": ", log.LstdFlags)
	logger.Printf("%s,%s,%s,%s\n", logData["updateFQDN"], logData["oldIP"], logData["newIP"], logData["msg"])

	return nil
}

func main() {

	var confFile string
	var currentIP string

	flag.StringVar(&confFile, "c", "none", "Specify Location of a config file for auth info and email recipients.")
	flag.Parse()

	confData, confErr := readConfig(confFile)
	if confErr != nil {
		log.Fatal(confErr)
	}

	if confData.UpdateParams.Debug {
		log.Printf("Updating Namecheap Dyn DNS for: %s.%s\n", confData.UpdateParams.Host, confData.UpdateParams.Domain)
	}

	currentIP, err := myIP("http://icanhazip.com/")
	if err != nil {
		currentIP, err = myIP("http://icmpzero.net/ip.php")
		if err != nil {
			if confData.UpdateParams.Debug {
				log.Printf("Failed to retrieve current public IP")
			}
			updateLogData := map[string]string{
				"oldIP":      "-",
				"newIP":      "-",
				"updateFQDN": fmt.Sprintf("%s.%s", confData.UpdateParams.Host, confData.UpdateParams.Domain),
				"msg":        "Failed to retrieve current public IP"}
			logErr := writeLog("updateLog.txt", "FAIL", updateLogData)
			if logErr != nil {
				log.Println("could not add log entry :\n", logErr)
			}
		}
	}

	if confData.UpdateParams.Debug {
		log.Printf("Current IP is %s \n", currentIP)
	}

	currentDNS, errDNS := currentDNS(fmt.Sprintf("%s.%s", confData.UpdateParams.Host, confData.UpdateParams.Domain))
	if errDNS != nil {

		if confData.UpdateParams.Debug {
			log.Printf("Could not find existing DNS entry")
		}
		updateLogData := map[string]string{
			"oldIP":      "-",
			"newIP":      currentIP,
			"updateFQDN": fmt.Sprintf("%s.%s", confData.UpdateParams.Host, confData.UpdateParams.Domain),
			"msg":        "Could not find existing DNS entry"}
		logErr := writeLog("updateLog.txt", "FAIL", updateLogData)
		if logErr != nil {
			log.Println("could not add log entry :\n", logErr)
		}
		os.Exit(0)
	}
	if confData.UpdateParams.Debug {
		log.Printf("Current DNS entry is %s\n", currentDNS)
	}

	if currentDNS[0] != currentIP {
		updateErr := updateSend(confData.UpdateParams.Host, confData.UpdateParams.Domain, confData.UpdateParams.Password, currentIP)
		if updateErr != nil {
			fmt.Println(updateErr)
		} else {
			if confData.UpdateParams.Debug {
				log.Printf("Update Successful\n\t%s\n\t%s\n\n", confData.UpdateParams.Host+"."+confData.UpdateParams.Domain, currentIP)
			}
			fmt.Printf("Update Successful\n\t%s\n\t%s\n\n", confData.UpdateParams.Host+"."+confData.UpdateParams.Domain, currentIP)
			if confData.UpdateParams.Log {
				updateLogData := map[string]string{
					"oldIP":      currentDNS[0],
					"newIP":      currentIP,
					"updateFQDN": fmt.Sprintf("%s.%s", confData.UpdateParams.Host, confData.UpdateParams.Domain),
					"msg":        "DNS record updated"}
				logErr := writeLog("updateLog.txt", "UPDATE", updateLogData)
				if logErr != nil {
					log.Println("could not add log entry :\n", logErr)
				}
			}
		}
	} else {
		if confData.UpdateParams.Debug {
			log.Printf("DNS and current IP match so no update necessary")
		}

		if true {
			updateLogData := map[string]string{
				"oldIP":      currentDNS[0],
				"newIP":      currentIP,
				"updateFQDN": fmt.Sprintf("%s.%s", confData.UpdateParams.Host, confData.UpdateParams.Domain),
				"msg":        "DNS and current IP match so no update necessary"}
			logErr := writeLog("updateLog.txt", "PASS", updateLogData)
			if logErr != nil {
				log.Println("could not add log entry :\n", logErr)
			}
		}
	}

}
