module github.com/noah-loop/backend/modules/llm

go 1.21

replace github.com/noah-loop/backend/shared => ../../shared

require (
	github.com/noah-loop/backend/shared v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid v1.4.0
	github.com/sashabaranov/go-openai v1.17.9
	github.com/anthropic/anthropic-sdk-go v0.1.0
	gorm.io/gorm v1.25.5
	github.com/google/wire v0.5.0
)
