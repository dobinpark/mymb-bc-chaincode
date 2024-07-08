package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type TokenERC1155Contract struct {
	contractapi.Contract
}

type Token1155 struct {
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

type User struct {
	UserId           string    `json:"userID"`
	NickName         string    `json:"nickName"`
	MymPoint         int64     `json:"mymPoint"`
	OwnedToken       []string  `json:"ownedToken"`
	BlockCreatedTime time.Time `json:"blockCreatedTime"`
}

const (
	tokenPrefix   = "token"
	balancePrefix = "balance"
)

// MintToken 토큰을 발행하는 함수
func (c *TokenERC1155Contract) MintToken(ctx contractapi.TransactionContextInterface, tokenNumber string, owner string,
	categoryCode string, fundingID string, ticketID string, tokenType string, sellStage string, imageURL string) (*Token1155, error) {

	user, err := c.GetUser(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get user information: %v", err)
	}

	if user.UserId == "" {
		return nil, fmt.Errorf("user %s does not exist", owner)
	}

	token := Token1155{
		TokenNumber:      tokenNumber,
		Owner:            owner,
		CategoryCode:     categoryCode,
		FundingID:        fundingID,
		TicketID:         ticketID,
		TokenType:        tokenType,
		SellStage:        sellStage,
		ImageURL:         imageURL,
		TokenCreatedTime: time.Now(),
	}

	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenNumber})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token: %v", err)
	}

	err = ctx.GetStub().PutState(tokenKey, tokenBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to put state: %v", err)
	}

	user.OwnedToken = append(user.OwnedToken, tokenNumber)

	userKey := owner
	userBytes, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user information: %v", err)
	}
	if err := ctx.GetStub().PutState(userKey, userBytes); err != nil {
		return nil, fmt.Errorf("failed to update user information: %v", err)
	}

	return &token, nil
}

// GetToken 해당 토큰을 조회하는 함수
func (c *TokenERC1155Contract) GetToken(ctx contractapi.TransactionContextInterface, tokenNumber string) (*Token1155, error) {

	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenNumber})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	tokenBytes, err := ctx.GetStub().GetState(tokenKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %v", err)
	}

	if tokenBytes == nil {
		return nil, fmt.Errorf("token %s does not exist", tokenNumber)
	}

	var token Token1155
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %v", err)
	}
	return &token, nil
}

// GetAllTokens 모든 토큰들을 조회하는 함수
func (c *TokenERC1155Contract) GetAllTokens(ctx contractapi.TransactionContextInterface) ([]Token1155, error) {

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(tokenPrefix, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to get state by partial composite key: %v", err)
	}
	defer resultsIterator.Close()

	var tokens []Token1155

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next query response: %v", err)
		}

		var token Token1155
		err = json.Unmarshal(queryResponse.Value, &token)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal token: %v", err)
		}

		tokens = append(tokens, token)
	}

	fmt.Printf("total: %d tokens\n", len(tokens))
	return tokens, nil
}

// GetTotalTokens 모든 토큰의 총 개수를 반환하는 함수
func (c *TokenERC1155Contract) GetTotalTokens(ctx contractapi.TransactionContextInterface) (int, error) {

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(tokenPrefix, []string{})
	if err != nil {
		return 0, fmt.Errorf("failed to get state by partial composite key: %v", err)
	}
	defer resultsIterator.Close()

	var totalCount int

	for resultsIterator.HasNext() {
		_, err := resultsIterator.Next()
		if err != nil {
			return 0, fmt.Errorf("failed to get next query response: %v", err)
		}
		totalCount++
	}

	fmt.Printf("total: %d tokens\n", totalCount)
	return totalCount, nil
}

// GetUserOwnedTokens 해당 유저가 가지고 있는 토큰들을 조회하는 함수
func (c *TokenERC1155Contract) GetUserOwnedTokens(ctx contractapi.TransactionContextInterface, nickName string) ([]*Token1155, error) {

	user, err := c.GetUser(ctx, nickName)
	if err != nil {
		return nil, fmt.Errorf("failed to get user information: %v", err)
	}

	if user.UserId == "" {
		return nil, fmt.Errorf("user %s does not exist", nickName)
	}

	var ownedTokens []*Token1155

	for _, tokenNumber := range user.OwnedToken {
		token, err := c.GetToken(ctx, tokenNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to get token %s: %v", tokenNumber, err)
		}
		ownedTokens = append(ownedTokens, token)
	}
	fmt.Printf("total: %d tokens\n", len(ownedTokens))
	return ownedTokens, nil
}

// UpdateSellStage sellStage 필드값을 변경하는 함수
func (c *TokenERC1155Contract) UpdateSellStage(ctx contractapi.TransactionContextInterface, tokenNumber string, newSellStage string) error {

	token, err := c.GetToken(ctx, tokenNumber)
	if err != nil {
		return fmt.Errorf("failed to get token: %v", err)
	}

	if token.TokenNumber == "" {
		return fmt.Errorf("token %s does not exist", tokenNumber)
	}

	token.SellStage = newSellStage

	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenNumber})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %v", err)
	}

	err = ctx.GetStub().PutState(tokenKey, tokenBytes)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}
	return nil
}

// TransferToken 지정된 토큰을 전송하는 함수
func (c *TokenERC1155Contract) TransferToken(ctx contractapi.TransactionContextInterface, from string, to string, tokenNumber string) error {

	fromUser, err := c.GetUser(ctx, from)
	if err != nil {
		return fmt.Errorf("failed to get sender information: %v", err)
	}

	if fromUser.UserId == "" {
		return fmt.Errorf("sender %s does not exist", from)
	}

	toUser, err := c.GetUser(ctx, to)
	if err != nil {
		return fmt.Errorf("failed to get receiver information: %v", err)
	}

	if toUser.UserId == "" {
		return fmt.Errorf("receiver %s does not exist", to)
	}

	found := false
	for _, t := range fromUser.OwnedToken {
		if t == tokenNumber {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("sender %s does not own the specified token %s", from, tokenNumber)
	}

	fromUser.OwnedToken = removeToken(fromUser.OwnedToken, tokenNumber)

	fromUserKey := from
	fromUserBytes, err := json.Marshal(fromUser)
	if err != nil {
		return fmt.Errorf("failed to marshal sender user: %v", err)
	}
	if err := ctx.GetStub().PutState(fromUserKey, fromUserBytes); err != nil {
		return fmt.Errorf("failed to update sender balance: %v", err)
	}

	toUser.OwnedToken = append(toUser.OwnedToken, tokenNumber)

	toUserKey := to
	toUserBytes, err := json.Marshal(toUser)
	if err != nil {
		return fmt.Errorf("failed to marshal receiver user: %v", err)
	}
	if err := ctx.GetStub().PutState(toUserKey, toUserBytes); err != nil {
		return fmt.Errorf("failed to update receiver balance: %v", err)
	}

	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenNumber})
	if err != nil {
		return fmt.Errorf("failed to create composite key for token: %v", err)
	}
	tokenBytes, err := ctx.GetStub().GetState(tokenKey)
	if err != nil {
		return fmt.Errorf("failed to get token information: %v", err)
	}
	if tokenBytes == nil {
		return fmt.Errorf("token %s does not exist", tokenNumber)
	}

	var token Token1155
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return fmt.Errorf("failed to unmarshal token: %v", err)
	}

	token.Owner = to

	tokenBytes, err = json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %v", err)
	}
	if err := ctx.GetStub().PutState(tokenKey, tokenBytes); err != nil {
		return fmt.Errorf("failed to update token owner: %v", err)
	}

	txID := ctx.GetStub().GetTxID()
	fmt.Printf("Transfer of token %s from %s to %s successfully recorded with transaction ID %s\n", tokenNumber, from, to, txID)

	return nil
}

// TransferAllTokens 해당 유저의 모든 토큰들을 전송하는 함수
func (c *TokenERC1155Contract) TransferAllTokens(ctx contractapi.TransactionContextInterface, from string, to string) error {

	fromUser, err := c.GetUser(ctx, from)
	if err != nil {
		return fmt.Errorf("failed to get user %s: %v", from, err)
	}

	if fromUser.UserId == "" {
		return fmt.Errorf("sender %s does not exist", from)
	}

	toUser, err := c.GetUser(ctx, to)
	if err != nil {
		return fmt.Errorf("failed to get user %s: %v", to, err)
	}

	if toUser.UserId == "" {
		return fmt.Errorf("receiver %s does not exist", to)
	}

	toUser.OwnedToken = append(toUser.OwnedToken, fromUser.OwnedToken...)
	fromUser.OwnedToken = []string{}

	fromUserBytes, err := json.Marshal(fromUser)
	if err != nil {
		return fmt.Errorf("failed to marshal user %s: %v", from, err)
	}
	err = ctx.GetStub().PutState(from, fromUserBytes)
	if err != nil {
		return fmt.Errorf("failed to put state for user %s: %v", from, err)
	}

	toUserBytes, err := json.Marshal(toUser)
	if err != nil {
		return fmt.Errorf("failed to marshal user %s: %v", to, err)
	}
	err = ctx.GetStub().PutState(to, toUserBytes)
	if err != nil {
		return fmt.Errorf("failed to put state for user %s: %v", to, err)
	}

	return nil
}

// DeleteTokens 지정된 토큰들을 삭제하는 함수
func (c *TokenERC1155Contract) DeleteTokens(ctx contractapi.TransactionContextInterface, nickName string, tokenNumbers []string) error {
	user, err := c.GetUser(ctx, nickName)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	if user.UserId == "" {
		return fmt.Errorf("user %s does not exist", nickName)
	}

	for _, tokenNumber := range tokenNumbers {
		user.OwnedToken = removeToken(user.OwnedToken, tokenNumber)

		tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenNumber})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}
		if err := ctx.GetStub().DelState(tokenKey); err != nil {
			return fmt.Errorf("failed to delete token: %v", err)
		}
	}

	userKey := user.NickName
	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %v", err)
	}
	if err := ctx.GetStub().PutState(userKey, userBytes); err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	return nil
}

// DeleteAllTokens 해당 유저가 가지고 있는 모든 토큰들을 삭제하는 함수
func (c *TokenERC1155Contract) DeleteAllTokens(ctx contractapi.TransactionContextInterface, nickName string) error {
	user, err := c.GetUser(ctx, nickName)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	if user.UserId == "" {
		return fmt.Errorf("user %s does not exist", nickName)
	}

	for _, tokenNumber := range user.OwnedToken {
		tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenNumber})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}
		if err := ctx.GetStub().DelState(tokenKey); err != nil {
			return fmt.Errorf("failed to delete token: %v", err)
		}
	}

	user.OwnedToken = []string{}

	userKey := user.NickName
	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %v", err)
	}
	if err := ctx.GetStub().PutState(userKey, userBytes); err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	return nil
}

// CreateUserBlock 유저 정보 블록을 생성하는 함수
func (c *TokenERC1155Contract) CreateUserBlock(ctx contractapi.TransactionContextInterface, userId string, nickName string, mymPoint int64, ownedToken []string) error {

	userBytes, err := ctx.GetStub().GetState(nickName)
	if err == nil && userBytes != nil {
		return fmt.Errorf("user %s already exists", nickName)
	}

	user := User{
		UserId:           userId,
		NickName:         nickName,
		MymPoint:         mymPoint,
		OwnedToken:       ownedToken,
		BlockCreatedTime: time.Now(),
	}

	userKey := nickName
	userBytes, err = json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user block: %v", err)
	}

	err = ctx.GetStub().PutState(userKey, userBytes)
	if err != nil {
		return fmt.Errorf("failed to put state for user block: %v", err)
	}
	return nil
}

// GetUser 해당 유저 정보를 조회하는 함수
func (c *TokenERC1155Contract) GetUser(ctx contractapi.TransactionContextInterface, nickName string) (*User, error) {

	userKey := nickName
	userBytes, err := ctx.GetStub().GetState(userKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read user block: %v", err)
	}

	if userBytes == nil {
		return &User{
			NickName:   nickName,
			OwnedToken: []string{},
		}, nil
	}

	var user User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user block: %v", err)
	}

	if user.OwnedToken == nil {
		user.OwnedToken = []string{}
	}

	return &user, nil
}

// GetAllUsers 모든 유저 정보를 조회하는 함수
func (c *TokenERC1155Contract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]User, error) {

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get state by range: %v", err)
	}
	defer resultsIterator.Close()

	var users []User

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next query response: %v", err)
		}

		var user User
		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal user: %v", err)
		}
		users = append(users, user)
	}
	fmt.Printf("total: %d users\n", len(users))
	return users, nil
}

// GetTotalUsers 모든 유저들의 total 값을 반환하는 함수
func (c *TokenERC1155Contract) GetTotalUsers(ctx contractapi.TransactionContextInterface) (int, error) {

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return 0, fmt.Errorf("failed to get state by range: %v", err)
	}
	defer resultsIterator.Close()

	var totalCount int

	for resultsIterator.HasNext() {
		_, err := resultsIterator.Next()
		if err != nil {
			return 0, fmt.Errorf("failed to get next query response: %v", err)
		}
		totalCount++
	}

	fmt.Printf("total: %d users\n", totalCount)
	return totalCount, nil
}

// DeleteUser 해당 닉네임을 가진 유저 블록을 삭제하는 함수
func (c *TokenERC1155Contract) DeleteUser(ctx contractapi.TransactionContextInterface, nickName string) error {
	userKey := nickName

	userBytes, err := ctx.GetStub().GetState(userKey)
	if err != nil {
		return fmt.Errorf("failed to read user block: %v", err)
	}
	if userBytes == nil {
		return fmt.Errorf("user with nickname %s does not exist", nickName)
	}

	err = ctx.GetStub().DelState(userKey)
	if err != nil {
		return fmt.Errorf("failed to delete user block: %v", err)
	}

	return nil
}

// DeleteAllUserBlocks 모든 유저 정보 블록을 삭제하는 함수
func (c *TokenERC1155Contract) DeleteAllUserBlocks(ctx contractapi.TransactionContextInterface) error {

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return fmt.Errorf("failed to get state by range: %v", err)
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return fmt.Errorf("failed to get next query response: %v", err)
		}

		err = ctx.GetStub().DelState(queryResponse.Key)
		if err != nil {
			return fmt.Errorf("failed to delete user block: %v", err)
		}
	}

	fmt.Println("All user blocks have been successfully deleted.")
	return nil
}

// UpdateMymPoint 커뮤니티 활동 포인트 적립하는 함수
func (c *TokenERC1155Contract) UpdateMymPoint(ctx contractapi.TransactionContextInterface, nickName string, delta int64) error {

	userKey := nickName
	userBytes, err := ctx.GetStub().GetState(userKey)
	if err != nil {
		return fmt.Errorf("failed to read user block: %v", err)
	}
	if userBytes == nil {
		return fmt.Errorf("user with nickname %s does not exist", nickName)
	}

	var user User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		return fmt.Errorf("failed to unmarshal user block: %v", err)
	}

	newMymPoint := user.MymPoint + delta
	if newMymPoint < 0 {
		return fmt.Errorf("MymPoint cannot be negative")
	}
	user.MymPoint = newMymPoint

	userBytes, err = json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal updated user block: %v", err)
	}

	err = ctx.GetStub().PutState(userKey, userBytes)
	if err != nil {
		return fmt.Errorf("failed to put state for updated user block: %v", err)
	}
	return nil
}

// 토큰 슬라이스에서 특정 토큰을 제거하는 도우미 함수
func removeToken(tokens []string, tokenNumber string) []string {
	var newTokens []string
	for _, token := range tokens {
		if token != tokenNumber {
			newTokens = append(newTokens, token)
		}
	}
	return newTokens
}

func main() {
	cc, err := contractapi.NewChaincode(new(TokenERC1155Contract))
	if err != nil {
		panic(err.Error())
	}
	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting TokenERC1155Contract chaincode: %s", err)
	}
}
