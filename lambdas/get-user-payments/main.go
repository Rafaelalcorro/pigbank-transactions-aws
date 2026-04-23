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
	userID := req.PathParameters["userId"]
	if userID == "" {
		return response(400, map[string]string{"error": "userId es requerido"}), nil
	}

	result, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("PAYMENT_TABLE")),
		IndexName:              aws.String("userId-index"),
		KeyConditionExpression: aws.String("userId = :userId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return response(500, map[string]string{"error": "error consultando DynamoDB: " + err.Error()}), nil
	}

	payments := []map[string]interface{}{}
	for _, item := range result.Items {
		payment := map[string]interface{}{}

		if v, ok := item["traceId"].(*types.AttributeValueMemberS); ok {
			payment["traceId"] = v.Value
		}
		if v, ok := item["status"].(*types.AttributeValueMemberS); ok {
			payment["status"] = v.Value
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

		payments = append(payments, payment)
	}

	return response(200, payments), nil
}

func main() {
	lambda.Start(handler)
}
