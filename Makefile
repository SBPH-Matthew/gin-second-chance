
fresh: 
	go run ./cmd/migrate fresh && go run ./cmd/seed 

start: 
	go run ./cmd/server