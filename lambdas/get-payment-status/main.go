package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var dynamoClient *dynamodb.Client

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("REGION")))
	dynamoClient = dynamodb.NewFromConfig(cfg)
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
	traceID := req.PathParameters["traceId"]
	if traceID == "" {
		return response(400, map[string]string{"error": "traceId es requerido"}), nil
	}

	result, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("PAYMENT_TABLE")),
		Key: map[string]types.AttributeValue{
			"traceId": &types.AttributeValueMemberS{Value: traceID},
		},
	})
	if err != nil {
		return response(500, map[string]string{"error": "error consultando DynamoDB: " + err.Error()}), nil
	}
	if result.Item == nil {
		return response(404, map[string]string{"error": "transacción no encontrada"}), nil
	}

	item := result.Item
	payment := map[string]interface{}{
		"traceId": traceID,
	}

	if v, ok := item["status"].(*types.AttributeValueMemberS); ok {
		payment["status"] = v.Value
	}
	if v, ok := item["userId"].(*types.AttributeValueMemberS); ok {
		payment["userId"] = v.Value
	}
	if v, ok := item["cardId"].(*types.AttributeValueMemberS); ok {
		payment["cardId"] = v.Value
	}
	if v, ok := item["timestamp"].(*types.AttributeValueMemberS); ok {
		payment["timestamp"] = v.Value
	}
	if v, ok := item["error"].(*types.AttributeValueMemberS); ok {
		payment["error"] = v.Value
	}
	if v, ok := item["service"].(*types.AttributeValueMemberS); ok {
		var service interface{}
		json.Unmarshal([]byte(v.Value), &service)
		payment["service"] = service
	}

	return response(200, payment), nil
}

func main() {
	lambda.Start(handler)
}
