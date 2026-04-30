module email

go 1.24

require (
	github.com/joho/godotenv v1.5.1
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/rs/zerolog v1.34.0
	pawly/pkg v0.0.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

replace pawly/pkg => ../pkg
