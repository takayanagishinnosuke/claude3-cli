package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"

	"github.com/joho/godotenv"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	filePathFlag := flag.String("img", "", "Please specify the file path. Example: -img <FilePath>")
	flag.Parse()

	if *filePathFlag != "" {
		filePath := *filePathFlag

		fmt.Println("Enter system prompt >>")
		systemPrompt, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read from stdin: %v", err)
		}
		systemPrompt = strings.TrimSpace(systemPrompt)

		fmt.Println("Enter your message >>")
		userMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read from stdin: %v", err)
		}
		userMessage = strings.TrimSpace(userMessage)

		response, err := claudeImgMessageCompletion(userMessage, systemPrompt, filePath)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("Response:", response)
		return
	}

	fmt.Println("Enter system prompt >>")
	systemPrompt, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read from stdin: %v", err)
	}
	systemPrompt = strings.TrimSpace(systemPrompt)

	for {
		fmt.Println("Enter your message >>")
		userMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read from stdin: %v", err)
		}
		userMessage = strings.TrimSpace(userMessage)

		err = claudeMessageStream(userMessage, systemPrompt)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	}
}

func claudeMessageStream(userMessage string, systemPrompt string) error {
	awsRegion, err := loadConfig()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// AWSのIAMユーザーとリージョンの設定
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion))
	if err != nil {
		return fmt.Errorf("error Config Load: %v", err)
	}

	// AWS Bedrockランタイムのクライアントを初期化
	brc := bedrockruntime.NewFromConfig(cfg)

	content := Content{Type: "text", Text: userMessage}

	msg := Message{Role: "user", Content: []Content{content}}

	payload := Request{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        512,
		System:           systemPrompt,
		Messages:         []Message{msg},
		Temperature:      0.5,
		TopP:             0.9,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %v", err)
	}

	output, err := brc.InvokeModelWithResponseStream(context.Background(), &bedrockruntime.InvokeModelWithResponseStreamInput{
		Body:        payloadBytes,
		ModelId:     aws.String("anthropic.claude-3-sonnet-20240229-v1:0"),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("error Claude response: %v", err)
	}

	err = processStreamingOutput(output, func(ctx context.Context, part []byte) error {
		if string(part) == "" {
			fmt.Print("\n")
			return nil
		}
		fmt.Print(string(part))
		return nil
	})

	if err != nil {
		log.Fatal("streaming output processing error: ", err)
	}

	return nil
}

type StreamingOutputHandler func(ctx context.Context, part []byte) error

func processStreamingOutput(output *bedrockruntime.InvokeModelWithResponseStreamOutput, handler StreamingOutputHandler) error {

	var combinedResult string

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:

			// fmt.Println("payload", string(v.Value.Bytes))
			var resp StreamResponse

			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				return err
			}

			err = handler(context.Background(), []byte(resp.Delta.Text))
			if err != nil {
				return err
			}

			combinedResult += resp.Delta.Text

		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)

		default:
			fmt.Println("union is nil or unknown type")
		}
	}

	return nil
}

func claudeImgMessageCompletion(userMessage string, systemPrompt string, filePath string) (string, error) {
	awsRegion, err := loadConfig()
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	// AWSのIAMユーザーとリージョンの設定
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion))
	if err != nil {
		return "", fmt.Errorf("error Config Load: %v", err)
	}

	// AWS Bedrockランタイムのクライアントを初期化
	brc := bedrockruntime.NewFromConfig(cfg)

	base64Image, err := encodeImageToBase64(filePath)
	if err != nil {
		log.Fatalf("Failed to encode image: %v", err)
	}

	contentImage := ContentImg{
		Type: "image",
		Source: &Source{
			Type:      "base64",
			MediaType: "image/jpeg",
			Data:      base64Image,
		},
	}

	contentText := ContentImg{
		Type: "text",
		Text: userMessage,
	}

	message := MessageImg{
		Role:    "user",
		Content: []ContentImg{contentImage, contentText},
	}

	request := RequestImg{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        1024,
		System:           systemPrompt,
		Messages:         []MessageImg{message},
	}

	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshalling payload: %v", err)
	}

	output, err := brc.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
		Body:        payloadBytes,
		ModelId:     aws.String("anthropic.claude-3-sonnet-20240229-v1:0"),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return "", fmt.Errorf("error Claude response: %v", err)
	}

	var resp Response

	err = json.Unmarshal(output.Body, &resp)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response: %v", err)
	}

	return resp.ContentItem[len(resp.ContentItem)-1].Text, nil
}

func encodeImageToBase64(filePath string) (string, error) {
	image, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(image), nil
}

func loadConfig() (string, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Could not read .env file")
	}
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	if awsAccessKey == "" {
		return "", fmt.Errorf("could not read AWS_ACCESS_KEY_ID")
	}
	awsSeacretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if awsSeacretAccessKey == "" {
		return "", fmt.Errorf("could not read AWS_SECRET_ACCESS_KEY")
	}
	awsRegion := os.Getenv("AWS_DEFAULT_REGION")
	if awsRegion == "" {
		return "", fmt.Errorf("could not read AWS_DEFAULT_REGION")
	}
	return awsRegion, nil
}
