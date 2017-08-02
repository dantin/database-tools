package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	host = flag.String("host", "", "Config server host")
	tier = flag.String("tier", "production", "Tier")
)

type ConfigServer struct {
	mysql MySqlConfig `json:"mysql"`
}

type MySqlConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Port     int    `json:"port"`
}

func getConfig() []byte {
	url := fmt.Sprintf("http://%s/master/api-%s.json", host, tier)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Fail to connect to config server: %s", url)
		os.Exit(1)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Fail to load configuration from %s", url)
		os.Exit(1)
	}

	return bytes.Replace(body, []byte("${ci.environment.slug}"), []byte(*tier), -1)
}

func main() {
	flag.Parse()

	fmt.Println(*tier)
}
