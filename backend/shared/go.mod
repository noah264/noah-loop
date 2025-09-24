module github.com/noah-loop/backend/shared

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid v1.4.0
	github.com/sirupsen/logrus v1.9.3
	gorm.io/gorm v1.25.5
	gorm.io/driver/postgres v1.5.4
	github.com/spf13/viper v1.17.0
	go.uber.org/zap v1.26.0
	github.com/prometheus/client_golang v1.17.0
	github.com/google/wire v0.5.0
	go.etcd.io/etcd/clientv3 v3.5.10
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/sdk v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.46.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.46.1
	go.opentelemetry.io/contrib/instrumentation/gorm.io/driver/postgres/otelpgx v0.46.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
	google.golang.org/grpc v1.59.0
	github.com/IBM/sarama v1.42.1
)
