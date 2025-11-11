module github.com/bagdasarian/checklist-app/api_service

go 1.24.1

require (
	github.com/bagdasarian/checklist-app/db_service v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/segmentio/kafka-go v0.4.49
	google.golang.org/genproto/googleapis/api v0.0.0-20251103181224-f26f9409b101
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

replace github.com/bagdasarian/checklist-app/db_service => ../db_service

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)
