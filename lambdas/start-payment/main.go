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

type SQSPayload struct {
	UserID    string  `json:"userId"`
	CardID    string  `json:"cardId"`
	Service   Service `json:"service"`
	TraceID   string  `json:"traceId"`
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
}

var dynamoClient *dynamodb.Client

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("REGION")))
	dynamoClient = dynamodb.NewFromConfig(cfg)
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	time.Sleep(5 * time.Second)

	for _, record := range sqsEvent.Records {
		var payload SQSPayload
		if err := json.Unmarshal([]byte(record.Body), &payload); err != nil {
			fmt.Printf("Error parseando mensaje: %v\n", err)
			continue
		}

		fmt.Printf("Procesando traceId: %s\n", payload.TraceID)

		serviceJSON, _ := json.Marshal(payload.Service)

		_, err := dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(os.Getenv("PAYMENT_TABLE")),
			Item: map[string]types.AttributeValue{
				"traceId":   &types.AttributeValueMemberS{Value: payload.TraceID},
				"userId":    &types.AttributeValueMemberS{Value: payload.UserID},
				"cardId":    &types.AttributeValueMemberS{Value: payload.CardID},
				"service":   &types.AttributeValueMemberS{Value: string(serviceJSON)},
				"status":    &types.AttributeValueMemberS{Value: "IN_PROGRESS"},
				"timestamp": &types.AttributeValueMemberS{Value: fmt.Sprintf("%d", time.Now().Unix())},
			},
		})
		if err != nil {
			fmt.Printf("Error guardando en DynamoDB: %v\n", err)
			return err
		}

		fmt.Printf("Estado IN_PROGRESS guardado para traceId: %s\n", payload.TraceID)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
