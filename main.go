// Reads the transactions from the API and
// insert them on an online MongoDB.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

// AllTransactions structure to receive one API call for all the transactions of an specific referenceId
type AllTransactions struct {
	ID     string  `json:"transactionId"`
	Amount string  `json:"amount"`
	Status float64 `json:"status"`
	Date   string  `json:"dtTransaction"`
}

// Configuration mongoDB and API authentication / PoS reference ID
type Configuration struct {
	MongoURI          string
	AuthenticationAPI string
	AuthenticationKey string
	ReferenceID       string
	LinkAPI           string
	DatabaseName      string
	CollectionName    string
}

// See README for instructions
func configs() (*Configuration, error) {
	cfg := new(Configuration)
	cfg.MongoURI = os.Getenv("MongoURI")
	cfg.AuthenticationAPI = os.Getenv("AuthenticationAPI")
	cfg.AuthenticationKey = os.Getenv("AuthenticationKey")
	cfg.ReferenceID = os.Getenv("ReferenceID")
	cfg.LinkAPI = os.Getenv("LinkAPI")
	cfg.DatabaseName = os.Getenv("DatabaseName")
	cfg.CollectionName = os.Getenv("CollectionName")
	return cfg, nil
}

func db(cfg *Configuration) (*mongo.Collection, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := mongo.Connect(ctx, cfg.MongoURI)
	if err != nil {
		return nil, err
	}
	collection := client.Database(cfg.DatabaseName).Collection(cfg.CollectionName)
	return collection, nil
}

func getTransactions(cfg *Configuration) ([]AllTransactions, error) {
	url := cfg.LinkAPI + cfg.ReferenceID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authenticationApi", cfg.AuthenticationAPI)
	req.Header.Set("authenticationKey", cfg.AuthenticationKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	transactions := []AllTransactions{}
	if err := json.NewDecoder(resp.Body).Decode(&transactions); err != nil {
		return nil, err
	}
	return transactions, nil
}

func insertTransactionDB(tsc []AllTransactions, clt *mongo.Collection, cfg *Configuration) error {
	for _, value := range tsc {
		var transaction map[string]interface{}
		client := &http.Client{}
		url := cfg.LinkAPI + "/" + value.ID
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("authenticationApi", cfg.AuthenticationAPI)
		req.Header.Set("authenticationKey", cfg.AuthenticationKey)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&transaction); err != nil {
			return err
		}
		filter := bson.M{"transactionId": string(value.ID)}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		count, err := clt.Count(ctx, filter)
		if err != nil {
			return err
		}
		if count == 0 {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_, err = clt.InsertOne(ctx, transaction)
			fmt.Printf("Transaction of ID %s inserted!\n", value.ID)
		}
	}
	return nil
}

func main() {
	// Get configuration parameters
	cfg, err := configs()
	if err != nil {
		log.Fatal(err)
	}
	// Config MongoDB and select collection
	collection, err := db(cfg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Fetching new transactions...\n")
	// Loop forever with 1 minute sleeps adding new transactions as they appear
	for {
		transactions, err := getTransactions(cfg)
		if err != nil {
			log.Fatal(err)
		}
		if err := insertTransactionDB(transactions, collection, cfg); err != nil {
			log.Fatal(err)
		}
		time.Sleep(1 * time.Minute)
	}
}
