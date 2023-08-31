package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
)

const URL = "https://go-challenge.skip.money"
const COLLECTION = "azuki"
const COLOR_GREEN = "\033[32m"
const COLOR_RED = "\033[31m"
const COLOR_RESET = "\033[0m"
const configKeyMaxThreads = "maxThreads"

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
var maxThreads = 1

type AttrsMap struct {
	attrCounts *SyncMap // attribute to number of occurrences (how many times does "green beret" occur for "hat"?)
}

type Token struct {
	id     int
	attrs  map[string]string
	rarity float64
}

type Collection struct {
	count int
	url   string
}

func getToken(tid int, colUrl string) *Token {
	url := fmt.Sprintf("%s/%s/%d.json", URL, colUrl, tid)
	res, err := http.Get(url)
	if err != nil {
		logger.Println(COLOR_RED, fmt.Sprintf("Error getting token %d :", tid), err, COLOR_RESET)
		return &Token{}
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Println(COLOR_RED, fmt.Sprintf("Error reading response for token %d :", tid), err, COLOR_RESET)
		return &Token{}
	}
	attrs := make(map[string]string)
	json.Unmarshal(body, &attrs)
	return &Token{
		id:    tid,
		attrs: attrs,
	}
}

func getTokens(col Collection) ([]*Token, AttrsMap) {
	tokens := make([]*Token, col.count)
	attrsMap := AttrsMap{
		attrCounts: NewSyncMap(),
	}

	var wg sync.WaitGroup

	numBuckets := maxThreads
	bucketSize := col.count / maxThreads
	overflow := len(tokens) % maxThreads

	for offset := 0; offset < numBuckets; offset++ {
		offset := offset
		wg.Add(1)
		go func() {
			defer wg.Done()

			// handle any overflow in the last "bucket"
			bucketWithOverflow := bucketSize
			if offset == numBuckets-1 {
				bucketWithOverflow += overflow
			}

			for bucket := 0; bucket < bucketWithOverflow; bucket++ {
				index := offset*bucketSize + bucket
				logger.Println(COLOR_GREEN, fmt.Sprintf("Getting token %d", index), COLOR_RESET)
				tokens[index] = getToken(index, col.url)

				// process to map
				for key, attr := range tokens[index].attrs {
					attrsMap.attrCounts.IncrementCount(key, attr)
				}
			}
		}()
	}
	wg.Wait()

	return tokens, attrsMap
}

func readConfig() {
	viper.SetConfigName("config") // name of config file (without an extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			panic(err)
		}
	} else {
		maxThreads = viper.GetInt(configKeyMaxThreads)
	}

	// the default value of 1 will be used if no config.yaml found
	logger.Println(COLOR_GREEN, fmt.Sprintf("using %d threads...", maxThreads))
}

func processRarity(token *Token, attrsMap AttrsMap) float64 {
	rarity := 0.0

	// rarity += 1 / (count_with_value * num_values_in_category

	for key, attr := range token.attrs {
		numEntries := attrsMap.attrCounts.NumEntries(key)
		attrOccurances, _ := attrsMap.attrCounts.Load(key, attr)

		rarity += 1.0 / float64(numEntries*attrOccurances)
	}

	return rarity
}

func processScores(tokens []*Token, attrsMap AttrsMap) {
	var wg sync.WaitGroup

	numBuckets := maxThreads
	bucketSize := len(tokens) / maxThreads
	overflow := len(tokens) % maxThreads

	for offset := 0; offset < numBuckets; offset++ {
		offset := offset
		wg.Add(1)
		go func() {

			defer wg.Done()

			// handle any overflow in the last "bucket"
			bucketWithOverflow := bucketSize
			if offset == numBuckets-1 {
				bucketWithOverflow += overflow - 1
			}

			for bucket := 0; bucket < bucketWithOverflow; bucket++ {
				index := offset*bucketSize + bucket

				// process score
				tokens[index].rarity = processRarity(tokens[index], attrsMap)
			}
		}()
	}

	wg.Wait()

}

func sortTokens(tokens []*Token) {
	// Sort by rarity, keeping original order or equal elements.
	sort.SliceStable(tokens, func(i, j int) bool {
		return tokens[i].rarity > tokens[j].rarity
	})
}

func main() {
	readConfig()

	azuki := Collection{
		count: 10000,
		url:   "azuki1",
	}

	tokens, attrsMap := getTokens(azuki)
	processScores(tokens, attrsMap)
	sortTokens(tokens) // could use some kind of parallel sorting algorithm here

	logger.Println(COLOR_GREEN, "Displaying Top 5 Tokens...", COLOR_RESET)
	for i := range tokens[:5] {
		logger.Println(COLOR_GREEN, fmt.Sprintf("ID: %d, Rarity: %.6f\n", tokens[i].id, tokens[i].rarity), COLOR_RESET)
	}
}

// plan
// as tokens are being viewed, build a traitmap of all available attrs
// as parallel as possible: foreach token -> calculate rarity score, yielding all scores
// use viper as config
