package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const (
	mongoURI          = "mongodb+srv://mymber:Alaql2022!@cluster-certifie.vkqpd9y.mongodb.net/?retryWrites=true&w=majority"
	database          = "MYMB_DB"
	fundingCollection = "fundingReferral"
	userCollection    = "user"
)

// BCUser 구조체 정의
type BCUser struct {
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

// User 구조체 정의
type User struct {
	ID                string    `bson:"_id"`
	Email             string    `bson:"email"`
	Password          string    `bson:"password"`
	TicketCount       int       `bson:"ticketCount"`
	ReferralCount     int       `bson:"referralCount"`
	NickName          string    `bson:"nickName"`
	InviterEmail      string    `bson:"inviterEmail"`
	MainCardId        string    `bson:"mainCardId"`
	MymId             string    `bson:"mymId"`
	IsEnterprise      bool      `bson:"isEnterprise"`
	CallNumber        string    `bson:"callNumber"`
	CountryCode       string    `bson:"countryCode"`
	BusinessNumber    string    `bson:"businessNumber"`
	FileName          string    `bson:"fileName"`
	UploadUrl         string    `bson:"uploadUrl"`
	TrustUsers        []string  `bson:"trustUsers"`
	TrustByUsers      []string  `bson:"trustByUsers"`
	IsIdentified      bool      `bson:"isIdentified"`
	CreatedAt         time.Time `bson:"createdAt"`
	DeletedAt         time.Time `bson:"deletedAt"`
	Name              string    `bson:"name"`
	IsCertificated    bool      `bson:"isCertificated"`
	BankAccount       string    `bson:"bankAccount"`
	BankName          string    `bson:"bankName"`
	AccountHolderName string    `bson:"accountHolderName"`
	PhoneNum          string    `bson:"phoneNum"`
}

// Function to execute the Docker command and get users
func getAllUsers() ([]BCUser, error) {
	cmd := exec.Command("docker", "exec", "cli", "peer", "chaincode", "query",
		"--tls", "--cafile", "/opt/home/managedblockchain-tls-chain.pem",
		"--channelID", "mychannel",
		"--name", "mycc",
		"-c", "{\"Args\":[\"GetAllUsers\"]}")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute chaincode: %v, output: %s", err, string(output))
	}

	var users []BCUser
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

	coll := client.Database(database).Collection(fundingCollection)
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

// Function to get a user's email by UID
func getUserEmailByID(uid string) (string, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return "", fmt.Errorf("failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	coll := client.Database(database).Collection(userCollection)
	var user User
	err = coll.FindOne(context.TODO(), bson.M{"_id": uid}).Decode(&user)
	if err != nil {
		return "", fmt.Errorf("failed to find user: %v", err)
	}

	return user.Email, nil
}

// Function to get funding referrals with email replacement
func getFundingReferralsWithEmails() ([]map[string]interface{}, error) {
	referrals, err := getFundingReferrals()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for _, referral := range referrals {
		fromEmail, err := getUserEmailByID(referral.ReferralFrom)
		if err != nil {
			return nil, err
		}
		toEmail, err := getUserEmailByID(referral.ReferralTo)
		if err != nil {
			return nil, err
		}
		result := map[string]interface{}{
			"referralPayback": referral.ReferralPayback,
			"fromEmail":       fromEmail,
			"toEmail":         toEmail,
		}
		results = append(results, result)
	}

	return results, nil
}

// Handler function to display users
func usersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := getAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Handler function to display tokens
func tokensHandler(w http.ResponseWriter, r *http.Request) {
	tokens, err := getAllTokens()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// Handler function to display funding referrals with emails
func fundingReferralsHandler(w http.ResponseWriter, r *http.Request) {
	referrals, err := getFundingReferralsWithEmails()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var customReferrals []string
	for _, referral := range referrals {
		customReferral := fmt.Sprintf(`{"referralPayback": %v, "%s", "%s"}`, referral["referralPayback"], referral["fromEmail"], referral["toEmail"])
		customReferrals = append(customReferrals, customReferral)
	}

	finalJSON := "[" + strings.Join(customReferrals, ",") + "]"

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(finalJSON))
}

func main() {
	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/tokens", tokensHandler)
	http.HandleFunc("/referrals", fundingReferralsHandler)
	fmt.Println("Server is listening on port 8090...")
	http.ListenAndServe("0.0.0.0:8090", nil)
}
