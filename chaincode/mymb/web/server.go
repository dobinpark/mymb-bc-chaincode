package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"html/template"
	"net/http"
	"os/exec"
	"time"
)

const (
	mongoURI   = "mongodb+srv://mymber:Alaql2022!@cluster-certifie.vkqpd9y.mongodb.net/?retryWrites=true&w=majority"
	database   = "MYMB_DB"
	collection = "fundingReferral"
)

// User 구조체 정의
type User struct {
	UserId           string   `json:"userID"`
	NickName         string   `json:"nickName"`
	MymPoint         int64    `json:"mymPoint"`
	OwnedToken       []string `json:"ownedToken"`
	BlockCreatedTime string   `json:"blockCreatedTime"`
}

// Token 구조체 정의
type Token struct {
	TokenNumber      string    `json:"tokenNumber"`
	Owner            string    `json:"owner"`
	CategoryCode     string    `json:"categoryCode"`
	FundingID        string    `json:"fundingID"`
	TicketID         string    `json:"ticketID"`
	TokenType        string    `json:"tokenType"`
	SellStage        string    `json:"sellStage"`
	ImageURL         string    `json:"imageURL"`
	TokenCreatedTime time.Time `json:"tokenCreatedTime"`
}

// FundingReferral 구조체 정의
type FundingReferral struct {
	FundingReferralId      string `json:"_id"`
	PayId                  string `json:"payId"`
	ReferralPayback        int    `json:"referralPayback"`
	ReferralFrom           string `json:"referralFrom"`
	ReferralTo             string `json:"referralTo"`
	IsBasePaymentCompleted bool   `json:"isBasePaymentCompleted"`
	IsPaybacked            bool   `json:"isPaybacked"`
}

// Function to execute the Docker command and get users
func getAllUsers() ([]User, error) {
	cmd := exec.Command("docker", "exec", "cli", "peer", "chaincode", "query",
		"--tls", "--cafile", "/opt/home/managedblockchain-tls-chain.pem",
		"--channelID", "mychannel",
		"--name", "mycc",
		"-c", "{\"Args\":[\"GetAllUsers\"]}")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute chaincode: %v, output: %s", err, string(output))
	}

	var users []User
	err = json.Unmarshal(output, &users)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	return users, nil
}

// Function to execute the Docker command and get tokens
func getAllTokens() ([]Token, error) {
	cmd := exec.Command("docker", "exec", "cli", "peer", "chaincode", "query",
		"--tls", "--cafile", "/opt/home/managedblockchain-tls-chain.pem",
		"--channelID", "mychannel",
		"--name", "mycc",
		"-c", "{\"Args\":[\"GetAllTokens\"]}")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute chaincode: %v, output: %s", err, string(output))
	}

	var tokens []Token
	err = json.Unmarshal(output, &tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	return tokens, nil
}

// MongoDB에서 FundingReferral 데이터를 가져오는 함수
func getFundingReferrals() ([]FundingReferral, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	coll := client.Database(database).Collection(collection)
	cur, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %v", err)
	}
	defer cur.Close(context.TODO())

	var referrals []FundingReferral
	for cur.Next(context.TODO()) {
		var referral FundingReferral
		err := cur.Decode(&referral)
		if err != nil {
			return nil, fmt.Errorf("failed to decode document: %v", err)
		}
		referrals = append(referrals, referral)
	}

	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return referrals, nil
}

// Handler function to display users
func usersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := getAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmplPath := "users.html"
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, users)
}

// Handler function to display tokens
func tokensHandler(w http.ResponseWriter, r *http.Request) {
	tokens, err := getAllTokens()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmplPath := "tokens.html"
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, tokens)
}

// Handler function to display funding referrals
func fundingReferralsHandler(w http.ResponseWriter, r *http.Request) {
	referrals, err := getFundingReferrals()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 필터링된 필드만 포함하도록 설정
	var filteredReferrals []FundingReferral
	for _, referral := range referrals {
		filteredReferrals = append(filteredReferrals, FundingReferral{
			ReferralPayback: referral.ReferralPayback,
			ReferralFrom:    referral.ReferralFrom,
			ReferralTo:      referral.ReferralTo,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filteredReferrals)
}

func main() {
	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/tokens", tokensHandler)
	http.HandleFunc("/referrals", fundingReferralsHandler)
	fmt.Println("Server is listening on port 8090...")
	http.ListenAndServe("0.0.0.0:8090", nil)
}
