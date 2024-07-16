package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"html/template"
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
	UserId           string   `bson:"userID"`
	NickName         string   `bson:"nickName"`
	MymPoint         int64    `bson:"mymPoint"`
	OwnedToken       []string `bson:"ownedToken"`
	BlockCreatedTime string   `bson:"blockCreatedTime"`
}

// Token 구조체 정의
type Token struct {
	TokenNumber      string    `bson:"tokenNumber"`
	Owner            string    `bson:"owner"`
	CategoryCode     string    `bson:"categoryCode"`
	FundingID        string    `bson:"fundingID"`
	TicketID         string    `bson:"ticketID"`
	TokenType        string    `bson:"tokenType"`
	SellStage        string    `bson:"sellStage"`
	ImageURL         string    `bson:"imageURL"`
	TokenCreatedTime time.Time `bson:"tokenCreatedTime"`
}

// FundingReferral 구조체 정의
type FundingReferral struct {
	FundingReferralId      primitive.ObjectID `bson:"_id,omitempty"`
	PayId                  string             `bson:"payId,omitempty"`
	ReferralPayback        int64              `bson:"referralPayback,omitempty"`
	ReferralFrom           string             `bson:"referralFrom,omitempty"`
	ReferralTo             string             `bson:"referralTo,omitempty"`
	IsBasePaymentCompleted bool               `bson:"isBasePaymentCompleted,omitempty"`
	IsPaybacked            bool               `bson:"isPaybacked,omitempty"`
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

// 은행 코드 매핑
var bankCodeMap = map[string]string{
	"001": "한국은행", "002": "산업은행", "003": "기업은행", "004": "국민은행", "005": "외환은행", "007": "수협은행",
	"008": "수출입은행", "011": "농협은행", "012": "농협회원조합", "020": "우리은행", "023": "SC제일은행", "026": "서울은행",
	"027": "한국씨티은행", "031": "대구은행", "032": "부산은행", "034": "광주은행", "035": "제주은행", "037": "전북은행",
	"039": "경남은행", "045": "새마을금고연합회", "048": "신협중앙회", "050": "상호저축은행", "051": "기타 외국계은행",
	"052": "모건스탠리은행", "054": "HSBC은행", "055": "도이치은행", "056": "알비에스피엘씨은행", "057": "제이피모간체이스은행",
	"058": "미즈호코퍼레이트은행", "059": "미쓰비시도쿄UFJ은행", "060": "BOA", "061": "비엔피파리바은행", "062": "중국공상은행",
	"063": "중국은행", "064": "산림조합", "065": "대화은행", "071": "우체국", "076": "신용보증기금", "077": "기술신용보증기금",
	"081": "하나은행", "088": "신한은행", "089": "케이뱅크", "090": "카카오뱅크", "092": "토스뱅크", "093": "한국주택금융공사",
	"094": "서울보증보험", "095": "경찰청", "099": "금융결제원", "209": "동양종합금융증권", "218": "현대증권", "230": "미래에셋증권",
	"238": "대우증권", "240": "삼성증권", "243": "한국투자증권", "247": "NH투자증권", "261": "교보증권", "262": "하이투자증권",
	"263": "에이치엠씨투자증권", "264": "키움증권", "265": "이트레이드증권", "266": "SK증권", "267": "대신증권",
	"268": "솔로몬투자증권", "269": "한화증권", "270": "하나대투증권", "278": "신한금융투자", "279": "동부증권",
	"280": "유진투자증권", "287": "메리츠증권", "289": "엔에이치투자증권", "290": "부국증권", "291": "신영증권",
	"292": "엘아이지투자증권",
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

// Function to get a user's email, bank account and bank name by UID
func getUserInfoByID(uid string) (string, string, string, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to connect to MongoDB: %v", err)
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
			return "", "", "", fmt.Errorf("failed to find user: %v", err)
		}
		return user.Email, user.BankAccount, user.BankName, nil
	}

	var user User
	err = coll.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to find user: %v", err)
	}

	return user.Email, user.BankAccount, user.BankName, nil
}

// Function to get funding referrals with email replacement
func getFundingReferralsWithEmails() ([]map[string]interface{}, error) {
	referrals, err := getFundingReferrals()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for _, referral := range referrals {
		fromEmail, _, _, err := getUserInfoByID(referral.ReferralFrom)
		if err != nil {
			return nil, err
		}
		toEmail, bankAccount, bankNameCode, err := getUserInfoByID(referral.ReferralTo)
		if err != nil {
			return nil, err
		}
		bankName, exists := bankCodeMap[bankNameCode]
		if !exists {
			bankName = bankNameCode // 코드가 없으면 그대로 사용
		}

		result := map[string]interface{}{
			"referralPayback": referral.ReferralPayback,
			"fromEmail":       fromEmail,
			"toEmail":         toEmail,
			"bankAccount":     bankAccount,
			"bankName":        bankName,
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

// Handler function to display funding referrals with emails and aggregated paybacks
func fundingReferralsHandler(w http.ResponseWriter, r *http.Request) {
	referrals, err := getFundingReferralsWithEmails()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 중복되는 toEmail에 대해 payback을 합산
	aggregatedPaybacks := make(map[string]map[string]interface{})
	for _, referral := range referrals {
		toEmail := referral["toEmail"].(string)
		referralPayback := referral["referralPayback"].(int)
		bankAccount := referral["bankAccount"].(string)
		bankName := referral["bankName"].(string)

		if _, exists := aggregatedPaybacks[toEmail]; !exists {
			aggregatedPaybacks[toEmail] = map[string]interface{}{
				"totalPayback": referralPayback,
				"toEmail":      toEmail,
				"bankAccount":  bankAccount,
				"bankName":     bankName,
			}
		} else {
			aggregatedPaybacks[toEmail]["totalPayback"] = aggregatedPaybacks[toEmail]["totalPayback"].(int) + referralPayback
		}
	}

	// 결과를 JSON으로 변환
	var customReferrals []string
	for _, data := range aggregatedPaybacks {
		customReferral := fmt.Sprintf(`{"totalPayback": %d, "toEmail": "%s", "bankAccount": "%s", "bankName": "%s"}`,
			data["totalPayback"].(int), data["toEmail"].(string), data["bankAccount"].(string), data["bankName"].(string))
		customReferrals = append(customReferrals, customReferral)
	}

	finalJSON := "[" + strings.Join(customReferrals, ",") + "]"

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(finalJSON))
}

func main() {
	// Initialize MongoDB client
	err := initMongoClient()
	if err != nil {
		fmt.Println("Failed to initialize MongoDB client:", err)
		return
	}

	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/tokens", tokensHandler)
	http.HandleFunc("/referrals", fundingReferralsHandler)
	fmt.Println("Server is listening on port 8090...")
	http.ListenAndServe("0.0.0.0:8090", nil)
}

// MongoDB 클라이언트 초기화 함수
func initMongoClient() error {
	var err error
	mongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}
	return nil
}

// MongoDB 클라이언트 (전역 변수로 재사용)
var mongoClient *mongo.Client

// MongoDB에서 FundingReferral 데이터를 가져오는 함수
func getFundingReferrals() ([]FundingReferral, error) {
	coll := mongoClient.Database(database).Collection(fundingCollection)
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
