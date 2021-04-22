package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/mbedford-stream/mbgofuncs/mbfile"
)

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

	// fmt.Println(currentIP)

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

	//"https://dynamicdns.park-your-domain.com/update?host=home&domain=misplaced-packets.net&password=6b0f73fe0f3142cba930ab8823809673&ip=97.81.124.44"

	return nil
}

type jsonConfig struct {
	UpdateParams struct {
		Domain   string `json:"domain"`
		Host     string `json:"host"`
		Password string `json:"password"`
	} `json:"updateParams"`
}

func readConfig(confFile string) (jsonConfig, error) {
	if !mbfile.FileExistsAndIsNotADirectory(confFile) {
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

func main() {

	var confFile string
	var currentIP string

	flag.StringVar(&confFile, "c", "none", "Specify Location of a config file for auth info and email recipients.")
	flag.Parse()

	confData, confErr := readConfig(confFile)
	if confErr != nil {
		log.Fatal(confErr)
	}

	fmt.Printf("Updating Namecheap Dyn DNS for: %s.%s\n", confData.UpdateParams.Host, confData.UpdateParams.Domain)

	currentIP, err := myIP("http://icanhazip.com/")
	if err != nil {
		currentIP, err = myIP("http://icmpzero.net/ip.php")
		if err != nil {
			log.Fatal("Failed to retrieve current public IP")
		}
	}

	fmt.Printf("Current IP is %s \n", currentIP)

	currentDNS, errDNS := currentDNS(confData.UpdateParams.Domain)
	if errDNS != nil {
		log.Fatal("Could not find existing DNS entry")
	}

	fmt.Printf("Current DNS entry is %s\n", currentDNS[0])

	fmt.Println(updateSend(confData.UpdateParams.Host, confData.UpdateParams.Domain, confData.UpdateParams.Password, currentIP))

}
