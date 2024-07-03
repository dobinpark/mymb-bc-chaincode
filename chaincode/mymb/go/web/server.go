package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"strings"
)

// User 구조체 정의 (contract 패키지에 정의된 구조체와 동일하게 맞춰야 합니다)
type User struct {
	UserId           string   `json:"userID"`
	NickName         string   `json:"nickName"`
	MymPoint         int64    `json:"mymPoint"`
	OwnedToken       []string `json:"ownedToken"`
	BlockCreatedTime string   `json:"blockCreatedTime"`
}

// Function to execute the Docker command and get users
func getAllUsers() ([]User, error) {

	caFilePath := "/opt/home/managedblockchain-tls-chain.pem"
	channelID := "mychannel"
	chaincodeName := "mycc"

	// fmt.Sprintf를 사용하여 포맷된 문자열 생성
	cmdStr := fmt.Sprintf("docker exec cli peer chaincode query --tls --cafile %s --channelID %s --name %s -c '{\"Args\":[\"GetAllUsers\"]}'", caFilePath, channelID, chaincodeName)
	cmdArgs := strings.Fields(cmdStr)

	// exec.Command를 사용하여 명령어 실행
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute chaincode: %v", err)
	}

	var users []User
	err = json.Unmarshal(output, &users)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	return users, nil
}

// Handler function to display users
func usersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := getAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("web/users.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, users)
}

func main() {
	http.HandleFunc("/users", usersHandler)
	fmt.Println("Server is listening on port 8090...")
	http.ListenAndServe(":8090", nil)
}
