package internal

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SqsClient struct {
	sqs   *sqs.SQS
	queue string
}

func NewSqsClient(region string, queue string) *SqsClient {
	return &SqsClient{
		sqs:   initSqs(region),
		queue: queue,
	}
}

func initSqs(region string) *sqs.SQS {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	return sqs.New(sess)
}

func (s *SqsClient) Charge(userId string, text string) error {
	_, err := s.sqs.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(fmt.Sprintf("%v:%v", userId, text)),
		QueueUrl:    aws.String(s.queue),
	})
	return err
}
