// chaincode/mymb/go/web/server.go
package main

import (
	"encoding/json"
	"fmt"
	contract "github.com/MYMB2022/mymb-bc-chaincode/chaincode/mymb/go"
	"html/template"
	"net/http"
	"os/exec"
	"path/filepath"
)

// Function to execute the chaincode command
func getAllUsers() ([]contract.User, error) {
	// token_contract의 실제 경로로 수정합니다.
	cmd := exec.Command(filepath.Join("..", "..", "token_contract"), "GetAllUsers")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute chaincode: %v", err)
	}

	var users []contract.User
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
