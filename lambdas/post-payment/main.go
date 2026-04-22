package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
)

type Service struct {
	ID            int     `json:"id"`
	Categoria     string  `json:"categoria"`
	Proveedor     string  `json:"proveedor"`
	Servicio      string  `json:"servicio"`
	Plan          string  `json:"plan"`
	PrecioMensual float64 `json:"precio_mensual"`
	Detalles      string  `json:"detalles"`
	Estado        string  `json:"estado"`
}

type PaymentRequest struct {
	CardID  string  `json:"cardId"`
	Service Service `json:"service"`
}

type SQSPayload struct {
	UserID    string  `json:"userId"`
	CardID    string  `json:"cardId"`
	Service   Service `json:"service"`
	TraceID   string  `json:"traceId"`
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
}

var (
	dynamoClient *dynamodb.Client
	sqsClient    *sqs.Client
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("REGION")))
	dynamoClient = dynamodb.NewFromConfig(cfg)
	sqsClient = sqs.NewFromConfig(cfg)
}

func response(status int, body any) events.APIGatewayProxyResponse {
	b, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var input PaymentRequest
	if err := json.Unmarshal([]byte(req.Body), &input); err != nil {
		return response(400, map[string]string{"error": "body inválido"}), nil
	}

	if input.CardID == "" {
		return response(400, map[string]string{"error": "cardId es requerido"}), nil
	}

	queryResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("CARD_TABLE")),
		KeyConditionExpression: aws.String("#uuid = :uuid"),
		ExpressionAttributeNames: map[string]string{
			"#uuid": "uuid",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid": &types.AttributeValueMemberS{Value: input.CardID},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return response(500, map[string]string{"error": "error DynamoDB: " + err.Error()}), nil
	}
	if len(queryResult.Items) == 0 {
		return response(404, map[string]string{"error": "tarjeta no encontrada"}), nil
	}

	card := queryResult.Items[0]

	statusAttr, ok := card["status"].(*types.AttributeValueMemberS)
	if !ok || (statusAttr.Value != "ACTIVE" && statusAttr.Value != "ACTIVATED") {
		return response(400, map[string]string{"error": "tarjeta no está activa"}), nil
	}

	userIDAttr, ok := card["user_id"].(*types.AttributeValueMemberS)
	if !ok {
		return response(500, map[string]string{"error": "tarjeta sin usuario asociado"}), nil
	}

	traceID := uuid.New().String()

	payload := SQSPayload{
		UserID:    userIDAttr.Value,
		CardID:    input.CardID,
		Service:   input.Service,
		TraceID:   traceID,
		Status:    "INITIAL",
		Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
	}

	payloadBytes, _ := json.Marshal(payload)

	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(os.Getenv("START_PAYMENT_SQS")),
		MessageBody: aws.String(string(payloadBytes)),
	})
	if err != nil {
		return response(500, map[string]string{"error": "error enviando a cola: " + err.Error()}), nil
	}

	return response(200, map[string]string{"traceId": traceID}), nil
}

func main() {
	lambda.Start(handler)
}
