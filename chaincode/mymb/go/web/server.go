package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"path/filepath"
)

// Token 구조체 정의
type Token1155 struct {
	TokenNumber      string `json:"tokenNumber"`
	Owner            string `json:"owner"`
	CategoryCode     string `json:"categoryCode"`
	FundingID        string `json:"fundingID"`
	TicketID         string `json:"ticketID"`
	TokenType        string `json:"tokenType"`
	SellStage        string `json:"sellStage"`
	ImageURL         string `json:"imageURL"`
	TokenCreatedTime string `json:"tokenCreatedTime"`
}

// User 구조체 정의
type User struct {
	UserId           string   `json:"userID"`
	NickName         string   `json:"nickName"`
	MymPoint         int64    `json:"mymPoint"`
	OwnedToken       []string `json:"ownedToken"`
	BlockCreatedTime string   `json:"blockCreatedTime"`
}

// Function to execute the Docker command and get users
func getAllUsers() ([]User, error) {
	// exec.Command를 사용하여 명령어 실행
	cmd := exec.Command("docker", "exec", "cli", "peer", "chaincode", "query",
		"--tls", "--cafile", "/opt/home/managedblockchain-tls-chain.pem",
		"--channelID", "mychannel",
		"--name", "mycc",
		"-c", "{\"Args\":[\"GetAllUsers\"]}")

	output, err := cmd.CombinedOutput() // CombinedOutput 사용하여 표준 출력과 표준 오류를 모두 캡처
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
func getAllTokens() ([]Token1155, error) {
	cmd := exec.Command("docker", "exec", "cli", "peer", "chaincode", "query",
		"--tls", "--cafile", "/opt/home/managedblockchain-tls-chain.pem",
		"--channelID", "mychannel",
		"--name", "mycc",
		"-c", "{\"Args\":[\"GetAllTokens\"]}")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute chaincode: %v, output: %s", err, string(output))
	}

	var tokens []Token1155
	err = json.Unmarshal(output, &tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	return tokens, nil
}

// Handler function to display users
func usersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := getAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 현재 파일의 디렉토리 경로를 기준으로 상대 경로를 설정합니다.
	tmplPath, err := filepath.Abs(filepath.Join("..", "web", "users.html"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get absolute path: %v", err), http.StatusInternalServerError)
		return
	}

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

	tmplPath, err := filepath.Abs(filepath.Join("web", "tokens.html"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get absolute path: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, tokens)
}

func main() {
	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/tokens", tokensHandler)
	fmt.Println("Server is listening on port 8090...")
	http.ListenAndServe("0.0.0.0:8090", nil)
}
