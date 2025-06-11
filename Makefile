run:
	go run ./cmd/fancaps-scraper $(ARGS)

build:
	go build -o fancaps-scraper-go ./cmd/fancaps-scraper