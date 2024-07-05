package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	BCUserId         string   `json:"BCUserId"`
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
	UserID            string    `json:"_id"`
	Email             string    `json:"email"`
	Password          string    `json:"password"`
	TicketCount       int       `json:"ticketCount"`
	ReferralCount     int       `json:"referralCount"`
	NickName          string    `json:"nickName"`
	InviterEmail      string    `json:"inviterEmail"`
	MainCardId        string    `json:"mainCardId"`
	MymId             string    `json:"mymId"`
	IsEnterprise      bool      `json:"isEnterprise"`
	CallNumber        string    `json:"callNumber"`
	CountryCode       string    `json:"countryCode"`
	BusinessNumber    string    `json:"businessNumber"`
	FileName          string    `json:"fileName"`
	UploadUrl         string    `json:"uploadUrl"`
	TrustUsers        []string  `json:"trustUsers"`
	TrustByUsers      []string  `json:"trustByUsers"`
	IsIdentified      bool      `json:"isIdentified"`
	CreatedAt         time.Time `json:"createdAt"`
	DeletedAt         time.Time `json:"deletedAt"`
	Name              string    `json:"name"`
	IsCertificated    bool      `json:"isCertificated"`
	BankAccount       string    `json:"bankAccount"`
	BankName          string    `json:"bankName"`
	AccountHolderName string    `json:"accountHolderName"`
	PhoneNum          string    `json:"phoneNum"`
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

	// Check if uid is a valid ObjectId
	objectID, err := primitive.ObjectIDFromHex(uid)
	if err != nil {
		// If not, treat uid as a string
		var user User
		err := coll.FindOne(context.TODO(), bson.M{"_id": uid}).Decode(&user)
		if err != nil {
			return "", fmt.Errorf("failed to find user: %v", err)
		}
		return user.Email, nil
	}

	var user User
	err = coll.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&user)
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
