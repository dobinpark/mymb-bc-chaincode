package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"time"
)

type TokenERC1155Contract struct {
	contractapi.Contract
}

type Token1155 struct {
	TokenID          string    `json:"TokenID"`
	CategoryCode     string    `json:"CategoryCode"`
	PollingResultID  string    `json:"PollingResultID"`
	TokenType        string    `json:"TokenType"`
	SellStage        string    `json:"sellStage"`
	TokenCreatedTime time.Time `json:"TokenCreatedTime"`
}

type User struct {
	NickName         string    `json:"NickName"`
	MymPoint         int64     `json:"MymPoint"`
	OwnedToken       []string  `json:"OwnedToken"`
	BlockCreatedTime time.Time `json:"BlockCreatedTime"`
}

type QueryResultToken struct {
	Key    string    `json:"Key"`
	Record Token1155 `json:"Record"`
}

type QueryResultUser struct {
	Key    string `json:"Key"`
	Record User   `json:"Record"`
}

const (
	tokenPrefix   = "token"
	balancePrefix = "balance"
)

func (c *TokenERC1155Contract) MintToken(ctx contractapi.TransactionContextInterface, tokenID string,
	categoryCode string, pollingResultID string, tokenType string, sellStage string) (*Token1155, error) {

	// Token 생성
	token := Token1155{
		TokenID:          tokenID,
		CategoryCode:     categoryCode,
		PollingResultID:  pollingResultID,
		TokenType:        tokenType,
		SellStage:        sellStage,
		TokenCreatedTime: time.Now(), // 현재 시간 사용
	}

	// TokenID, Token 저장
	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenID})
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

	// 사용자의 ownedToken 필드에 토큰 추가(닉네임을 고정하여 사용)
	nickName := "(주)밈비" // 닉네임을 고정
	user, err := c.GetUser(ctx, nickName)
	if err != nil {
		return nil, fmt.Errorf("failed to get user information: %v", err)
	}

	user.OwnedToken = append(user.OwnedToken, tokenID)

	userKey := nickName
	userBytes, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user information: %v", err)
	}
	if err := ctx.GetStub().PutState(userKey, userBytes); err != nil {
		return nil, fmt.Errorf("failed to update user information: %v", err)
	}

	return &token, nil
}

// sellStage 값을 변경하는 함수.
func (c *TokenERC1155Contract) UpdateSellStage(ctx contractapi.TransactionContextInterface, tokenID string, newSellStage string) error {

	// 토큰 조회
	token, err := c.GetToken(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to get token: %v", err)
	}

	// sellStage 필드 업데이트
	token.SellStage = newSellStage

	// 토큰 정보 업데이트
	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenID})
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

func (c *TokenERC1155Contract) CreateUserBlock(ctx contractapi.TransactionContextInterface,
	nickname string, mymPoint int64, ownedToken []string) error {

	// User 생성
	user := User{
		NickName:         nickname,
		MymPoint:         mymPoint,
		OwnedToken:       ownedToken,
		BlockCreatedTime: time.Now(),
	}

	// User 블록 저장
	userKey := nickname // 닉네임을 키로 사용
	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marsshal user block: %v", err)
	}

	err = ctx.GetStub().PutState(userKey, userBytes)
	if err != nil {
		return fmt.Errorf("failed to put state for user block: %v", err)
	}
	return nil
}

func (c *TokenERC1155Contract) UpdateMymPoint(ctx contractapi.TransactionContextInterface, nickName string, delta int64) error {

	// 기존 유저 정보 가져오기
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

	// MymPoint 업데이트
	user.MymPoint += delta

	// 업데이트된 유저 정보 저장
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

func (c *TokenERC1155Contract) GetToken(ctx contractapi.TransactionContextInterface, tokenID string) (*Token1155, error) {

	// 토큰 ID를 사용하여 토큰 키 생성
	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	// 토큰 상태 조회
	tokenBytes, err := ctx.GetStub().GetState(tokenKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %v", err)
	}

	// 조회된 토큰이 없으면 에러 반환
	if tokenBytes == nil {
		return nil, fmt.Errorf("token with ID %s does not exist", tokenID)
	}

	// 조회된 토큰을 구조체로 변환하여 반환
	var token Token1155
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %v", err)
	}
	return &token, nil
}

func (c *TokenERC1155Contract) GetAllTokens(ctx contractapi.TransactionContextInterface) ([]QueryResultToken, error) {

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(tokenPrefix, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to get state by partial composite key: %v", err)
	}
	defer resultsIterator.Close()

	var results []QueryResultToken

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

		results = append(results, QueryResultToken{
			Key:    queryResponse.Key,
			Record: token,
		})
	}
	return results, nil
}

func (c *TokenERC1155Contract) GetUserOwnedTokens(ctx contractapi.TransactionContextInterface, nickName string) ([]*Token1155, error) {
	// 사용자 정보 조회
	user, err := c.GetUser(ctx, nickName)
	if err != nil {
		return nil, fmt.Errorf("failed to get user information: %v", err)
	}

	var ownedTokens []*Token1155

	// 사용자가 소유한 각 토큰 ID에 대해 토큰 정보 조회
	for _, tokenID := range user.OwnedToken {
		token, err := c.GetToken(ctx, tokenID)
		if err != nil {
			return nil, fmt.Errorf("failed to get token %s: %v", tokenID, err)
		}
		ownedTokens = append(ownedTokens, token)
	}
	return ownedTokens, nil
}

func (c *TokenERC1155Contract) GetUser(ctx contractapi.TransactionContextInterface, nickName string) (*User, error) {

	userKey := nickName // 닉네임을 키로 사용
	userBytes, err := ctx.GetStub().GetState(userKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read user block: %v", err)
	}
	if userBytes == nil {
		return nil, fmt.Errorf("user with nickname %s does not exist", nickName)
	}

	var user User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user block: %v", err)
	}
	return &user, nil
}

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
	return users, nil
}

// 특정 사용자가 다른 사용자에게 토큰을 전송하는 함수
func (c *TokenERC1155Contract) TransferToken(ctx contractapi.TransactionContextInterface, from string, to string, tokenID []string) error {

	// 송신자와 수신자의 정보 가져오기
	fromUser, err := c.GetUser(ctx, from)
	if err != nil {
		return fmt.Errorf("failed to get sender information: %v", err)
	}
	toUser, err := c.GetUser(ctx, to)
	if err != nil {
		return fmt.Errorf("failed to get receiver information: %v", err)
	}

	// 송신자의 토큰 선택
	fromUserTokens, err := c.GetUserOwnedTokens(ctx, from)
	if err != nil {
		return fmt.Errorf("failed to get sender's tokens: %v", err)
	}

	// 송신자가 보유한 토큰들 중에 전송할 토큰들을 선택
	var transferTokens []*Token1155
	for _, fromUserToken := range fromUserTokens {
		for _, tokenID := range tokenID {
			if fromUserToken.TokenID == tokenID {
				transferTokens = append(transferTokens, fromUserToken)
				break
			}
		}
	}

	// 송신자가 전송할 토큰이 없는 경우 오류 반환
	if len(transferTokens) != len(tokenID) {
		return fmt.Errorf("sender %s does not own all specified tokens", from)
	}

	// 송신자의 토큰 잔고 갱신
	for _, token := range transferTokens {
		fromUser.OwnedToken = removeToken(fromUser.OwnedToken, token.TokenID)
	}

	// 송신자 정보 업데이트
	fromUserKey := from // 닉네임을 사용하여 사용자 키 생성
	fromUserBytes, err := json.Marshal(fromUser)
	if err != nil {
		return fmt.Errorf("failed to marshal sender user: %v", err)
	}
	if err := ctx.GetStub().PutState(fromUserKey, fromUserBytes); err != nil {
		return fmt.Errorf("failed to update sender balance: %v", err)
	}

	// 수신자의 토큰 잔고 갱신
	toUser.OwnedToken = append(toUser.OwnedToken, tokenID...)

	// 수신자 정보 업데이트
	toUserKey := to // 닉네임을 사용하여 사용자 키 생성
	toUserBytes, err := json.Marshal(toUser)
	if err != nil {
		return fmt.Errorf("failed to marshal receiver user: %v", err)
	}
	if err := ctx.GetStub().PutState(toUserKey, toUserBytes); err != nil {
		return fmt.Errorf("failed to update receiver balance: %v", err)
	}
	return nil
}

// 토큰 슬라이스에서 특정 토큰을 제거하는 도우미 함수
func removeToken(tokens []string, tokenID string) []string {
	var newTokens []string
	for _, token := range tokens {
		if token != tokenID {
			newTokens = append(newTokens, token)
		}
	}
	return newTokens
}

func main() {
	// The main function is not required for Hyperledger Fabric chaincode
	// It's here only for demonstration purposes
	cc, err := contractapi.NewChaincode(new(TokenERC1155Contract))
	if err != nil {
		panic(err.Error())
	}
	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting TokenERC1155Contract chaincode: %s", err)
	}
}
