package internal

import (
	"errors"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	ErrBalanceParsing = errors.New("can't read balance")
	ErrBalanceFetch   = errors.New("can't fetch balance")
	ErrTopup          = errors.New("can't topup balance")
	ErrInsufficient   = errors.New("balance is not sufficient")
	ErrCharge         = errors.New("can't charge balance")
)

type BalanceClient struct {
	table string
	db    *dynamodb.DynamoDB
}

func NewBalanceClient(
	table string,
	region string,
) *BalanceClient {
	db := initDb(region)
	return &BalanceClient{
		table: table,
		db:    db,
	}
}

func FromDynamoDb(
	table string,
	db *dynamodb.DynamoDB,
) *BalanceClient {
	return &BalanceClient{
		table: table,
		db:    db,
	}
}

func initDb(region string) *dynamodb.DynamoDB {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)},
	))
	return dynamodb.New(sess)
}

func (c *BalanceClient) Balance(userId string) (int, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(c.table),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(userId),
			},
		},
	}
	result, err := c.db.GetItem(input)
	if err != nil {
		return 0, ErrBalanceFetch
	}
	balanceAttr := result.Item["balance"]
	if balanceAttr == nil {
		return 0, nil
	}
	balance, err := strconv.Atoi(*balanceAttr.N)
	if err != nil {
		return 0, ErrBalanceParsing
	}
	return balance, nil
}

func (c *BalanceClient) Topup(userId string, value int) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(c.table),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(userId),
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#B": aws.String("balance"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":inc": {
				N: aws.String(strconv.Itoa(value)),
			},
		},
		UpdateExpression: aws.String("ADD #B :inc"),
	}

	_, err := c.db.UpdateItem(input)
	if err != nil {
		return ErrTopup
	}
	return nil
}

func (c *BalanceClient) Charge(userId string, value int) error {
	updateInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(c.table),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(userId),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":val": {
				N: aws.String(strconv.Itoa(value)),
			},
		},
		ConditionExpression: aws.String("balance >= :val"),
		UpdateExpression:    aws.String("set balance = balance - :val"),
	}
	_, err := c.db.UpdateItem(updateInput)
	if err != nil {
		var conditionalCheckFailedException *dynamodb.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedException) {
			return ErrInsufficient
		}
		return ErrCharge
	}
	return nil
}
