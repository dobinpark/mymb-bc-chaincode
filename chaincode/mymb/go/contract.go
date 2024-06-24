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
	TokenNumber      string    `json:"TokenNumber"`
	Owner            string    `json:"Owner"`
	CategoryCode     string    `json:"CategoryCode"`
	FundingID        string    `json:"FundingID"`
	TicketID         string    `json:"TicketID"`
	TokenType        string    `json:"TokenType"`
	SellStage        string    `json:"sellStage"`
	ImageURL         string    `json:"imageURL"`
	TokenCreatedTime time.Time `json:"TokenCreatedTime"`
}

type User struct {
	UserId           string    `json:"UserID"`
	NickName         string    `json:"NickName"`
	MymPoint         int64     `json:"MymPoint"`
	OwnedToken       []string  `json:"OwnedToken"`
	BlockCreatedTime time.Time `json:"BlockCreatedTime"`
}

type TransferRequest struct {
	From         string   `json:"from"`
	To           string   `json:"to"`
	TokenNumbers []string `json:"tokenNumbers"`
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

// MintToken 토큰을 발행하는 함수
func (c *TokenERC1155Contract) MintToken(ctx contractapi.TransactionContextInterface, tokenNumber string, owner string,
	categoryCode string, fundingID string, ticketID string, tokenType string, sellStage string, imageURL string) (*Token1155, error) {

	// Token 생성
	token := Token1155{
		TokenNumber:      tokenNumber,
		Owner:            owner,
		CategoryCode:     categoryCode,
		FundingID:        fundingID,
		TicketID:         ticketID,
		TokenType:        tokenType,
		SellStage:        sellStage,
		ImageURL:         imageURL,
		TokenCreatedTime: time.Now(), // 현재 시간 사용
	}

	// TokenNumber를 사용하여 Composite Key 생성 및 Token 저장
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

	// 사용자의 ownedToken 필드에 토큰 추가
	user, err := c.GetUser(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get user information: %v", err)
	}

	// 사용자 정보 업데이트 전 출력
	fmt.Printf("User before update: %+v\n", user)

	user.OwnedToken = append(user.OwnedToken, tokenNumber)

	// 사용자 정보 업데이트 후 출력
	fmt.Printf("User after update: %+v\n", user)

	userKey := owner // 닉네임을 키로 사용
	userBytes, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user information: %v", err)
	}
	if err := ctx.GetStub().PutState(userKey, userBytes); err != nil {
		return nil, fmt.Errorf("failed to update user information: %v", err)
	}

	// 상태 업데이트 확인
	updatedUserBytes, err := ctx.GetStub().GetState(userKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user information: %v", err)
	}
	var updatedUser User
	if err := json.Unmarshal(updatedUserBytes, &updatedUser); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated user information: %v", err)
	}
	fmt.Printf("Updated user state: %+v\n", updatedUser)

	return &token, nil
}

// GetToken 해당 토큰을 조회하는 함수
func (c *TokenERC1155Contract) GetToken(ctx contractapi.TransactionContextInterface, tokenNumber string) (*Token1155, error) {

	// 토큰 Number를 사용하여 토큰 키 생성
	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenNumber})
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
		return nil, fmt.Errorf("token with Number %s does not exist", tokenNumber)
	}

	// 조회된 토큰을 구조체로 변환하여 반환
	var token Token1155
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %v", err)
	}
	return &token, nil
}

// GetAllTokens 모든 토큰들을 조회하는 함수
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
	// 총 개수를 로그에 출력
	fmt.Printf("total: %d tokens\n", len(results))
	return results, nil
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

	// 총 개수를 로그에 출력
	fmt.Printf("total: %d tokens\n", totalCount)
	return totalCount, nil
}

// GetUserOwnedTokens 해당 유저가 가지고 있는 토큰들을 조회하는 함수
func (c *TokenERC1155Contract) GetUserOwnedTokens(ctx contractapi.TransactionContextInterface, nickName string) ([]*Token1155, error) {

	// 사용자 정보 조회
	user, err := c.GetUser(ctx, nickName)
	if err != nil {
		return nil, fmt.Errorf("failed to get user information: %v", err)
	}

	var ownedTokens []*Token1155

	// 사용자가 소유한 각 토큰 Number에 대해 토큰 정보 조회
	for _, tokenNumber := range user.OwnedToken {
		token, err := c.GetToken(ctx, tokenNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to get token %s: %v", tokenNumber, err)
		}
		ownedTokens = append(ownedTokens, token)
	}
	// 총 개수를 로그에 출력
	fmt.Printf("total: %d tokens\n", len(ownedTokens))
	return ownedTokens, nil
}

// UpdateSellStage sellStage 필드값을 변경하는 함수
func (c *TokenERC1155Contract) UpdateSellStage(ctx contractapi.TransactionContextInterface, tokenNumber string, newSellStage string) error {

	// 토큰 조회
	token, err := c.GetToken(ctx, tokenNumber)
	if err != nil {
		return fmt.Errorf("failed to get token: %v", err)
	}

	// sellStage 필드 업데이트
	token.SellStage = newSellStage

	// 토큰 정보 업데이트
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

	// 송신자와 수신자의 정보 가져오기
	fromUser, err := c.GetUser(ctx, from)
	if err != nil {
		return fmt.Errorf("failed to get sender information: %v", err)
	}
	toUser, err := c.GetUser(ctx, to)
	if err != nil {
		return fmt.Errorf("failed to get receiver information: %v", err)
	}

	// 송신자가 보유한 토큰 중에서 전송할 토큰을 찾기
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

	// 송신자의 토큰 잔고 갱신
	fromUser.OwnedToken = removeToken(fromUser.OwnedToken, tokenNumber)

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
	toUser.OwnedToken = append(toUser.OwnedToken, tokenNumber)

	// 수신자 정보 업데이트
	toUserKey := to // 닉네임을 사용하여 사용자 키 생성
	toUserBytes, err := json.Marshal(toUser)
	if err != nil {
		return fmt.Errorf("failed to marshal receiver user: %v", err)
	}
	if err := ctx.GetStub().PutState(toUserKey, toUserBytes); err != nil {
		return fmt.Errorf("failed to update receiver balance: %v", err)
	}

	// 토큰 정보 가져오기
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

	// 토큰의 소유자 업데이트
	var token Token1155
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return fmt.Errorf("failed to unmarshal token: %v", err)
	}

	// 토큰의 Owner 필드를 'to'로 변경
	token.Owner = to

	// 업데이트된 토큰 정보 저장
	tokenBytes, err = json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %v", err)
	}
	if err := ctx.GetStub().PutState(tokenKey, tokenBytes); err != nil {
		return fmt.Errorf("failed to update token owner: %v", err)
	}

	// 트랜잭션 성공적으로 기록 확인
	txID := ctx.GetStub().GetTxID()
	fmt.Printf("Transfer of token %s from %s to %s successfully recorded with transaction ID %s\n", tokenNumber, from, to, txID)

	return nil
}

// TransferAllTokens 해당 유저의 모든 토큰들을 전송하는 함수
func (c *TokenERC1155Contract) TransferAllTokens(ctx contractapi.TransactionContextInterface, from string, to string) error {

	// 발신자 사용자 정보 조회
	fromUser, err := c.GetUser(ctx, from)
	if err != nil {
		return fmt.Errorf("failed to get user %s: %v", from, err)
	}

	// 수신자 사용자 정보 조회
	toUser, err := c.GetUser(ctx, to)
	if err != nil {
		return fmt.Errorf("failed to get user %s: %v", to, err)
	}

	// 모든 토큰을 수신자에게 전송
	toUser.OwnedToken = append(toUser.OwnedToken, fromUser.OwnedToken...)
	fromUser.OwnedToken = []string{}

	// 발신자 사용자 정보 업데이트
	fromUserBytes, err := json.Marshal(fromUser)
	if err != nil {
		return fmt.Errorf("failed to marshal user %s: %v", from, err)
	}
	err = ctx.GetStub().PutState(from, fromUserBytes)
	if err != nil {
		return fmt.Errorf("failed to put state for user %s: %v", from, err)
	}

	// 수신자 사용자 정보 업데이트
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
	// 사용자의 토큰 목록 가져오기
	user, err := c.GetUser(ctx, nickName)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	// 토큰 목록에서 지정된 토큰들 제거
	for _, tokenNumbers := range tokenNumbers {
		user.OwnedToken = removeToken(user.OwnedToken, tokenNumbers)

		// 체인코드 상태에서 토큰 삭제
		tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenNumbers})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}
		if err := ctx.GetStub().DelState(tokenKey); err != nil {
			return fmt.Errorf("failed to delete token: %v", err)
		}
	}

	// 업데이트된 사용자 정보 저장
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
	// 사용자의 토큰 목록 가져오기
	user, err := c.GetUser(ctx, nickName)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	// 모든 토큰 삭제
	for _, tokenNumber := range user.OwnedToken {
		// 체인코드 상태에서 토큰 삭제
		tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenNumber})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}
		if err := ctx.GetStub().DelState(tokenKey); err != nil {
			return fmt.Errorf("failed to delete token: %v", err)
		}
	}

	// 사용자의 OwnedToken 필드 초기화
	user.OwnedToken = []string{}

	// 업데이트된 사용자 정보 저장
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
func (c *TokenERC1155Contract) CreateUserBlock(ctx contractapi.TransactionContextInterface, userId string, nickname string, mymPoint int64, ownedToken []string) error {

	// User 생성
	user := User{
		UserId:           userId,
		NickName:         nickname,
		MymPoint:         mymPoint,
		OwnedToken:       ownedToken,
		BlockCreatedTime: time.Now(),
	}

	// User 블록 저장
	userKey := userId // 유저ID를 키로 사용
	userBytes, err := json.Marshal(user)
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

	// 닉네임을 UTF-8로 인코딩
	userKey := nickName
	userBytes, err := ctx.GetStub().GetState(userKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read user block: %v", err)
	}

	if userBytes == nil {
		// 존재하지 않는 사용자인 경우 빈 User 객체를 반환
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

	// OwnedToken 필드가 nil인 경우 빈 배열로 초기화
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
	// 총 개수를 로그에 출력
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

	// 총 개수를 로그에 출력
	fmt.Printf("total: %d users\n", totalCount)
	return totalCount, nil
}

// DeleteUser 해당 닉네임을 가진 유저 블록을 삭제하는 함수
func (c *TokenERC1155Contract) DeleteUser(ctx contractapi.TransactionContextInterface, nickName string) error {
	// 닉네임을 키로 사용
	userKey := nickName

	// 유저가 존재하는지 확인
	userBytes, err := ctx.GetStub().GetState(userKey)
	if err != nil {
		return fmt.Errorf("failed to read user block: %v", err)
	}
	if userBytes == nil {
		// 존재하지 않는 사용자인 경우 빈 User 객체를 반환
		return fmt.Errorf("user with nickname %s does not exist", nickName)
	}

	// 유저 블록 삭제
	err = ctx.GetStub().DelState(userKey)
	if err != nil {
		return fmt.Errorf("failed to delete user block: %v", err)
	}

	return nil
}

// DeleteAllUserBlocks 모든 유저 정보 블록을 삭제하는 함수
func (c *TokenERC1155Contract) DeleteAllUserBlocks(ctx contractapi.TransactionContextInterface) error {

	// 전체 유저 상태 조회
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return fmt.Errorf("failed to get state by range: %v", err)
	}
	defer resultsIterator.Close()

	// 각 유저 상태 삭제
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

	// MymPoint 업데이트(음수 방지)
	newMymPoint := user.MymPoint + delta
	if newMymPoint < 0 {
		return fmt.Errorf("MymPoint cannot be negative")
	}
	user.MymPoint = newMymPoint

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
