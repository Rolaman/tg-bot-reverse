package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"
	"tg-bot-balance/internal"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestBalance(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "amazon/dynamodb-local:1.13.6",
			ExposedPorts: []string{"8000/tcp"},
			WaitingFor:   wait.ForLog("Initializing DynamoDB Local with the following configuration"),
		},
		Started: true,
	}
	dynamoDbContainer, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	defer dynamoDbContainer.Terminate(ctx)

	endpoint, err := dynamoDbContainer.Endpoint(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	awsSession := session.Must(session.NewSession(&aws.Config{
		Region:   aws.String("us-west-1"),
		Endpoint: aws.String("http://" + endpoint),
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}))

	dynamoDb := dynamodb.New(awsSession)
	tableName := "balance_table"
	_, err = dynamoDb.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	client := internal.FromDynamoDb(tableName, dynamoDb)
	err = client.Topup("1", 100)
	if err != nil {
		t.Fatal(err)
	}

	balance, err := client.Balance("1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, balance, 100, "balance should be equal 100")

	balance, err = client.Balance("2")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, balance, 0, "balance should be equal 0")
}
