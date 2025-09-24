module github.com/noah-loop/backend/modules/agent

go 1.21

replace github.com/noah-loop/backend/shared => ../../shared

require (
	github.com/noah-loop/backend/shared v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid v1.4.0
	gorm.io/gorm v1.25.5
	github.com/fsnotify/fsnotify v1.7.0
)
