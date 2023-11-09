To run:
go run src/app.go

To containerize (you can name the container what you like and also define your own tags): 
docker build -t cs302-project/workshop:1.0 ./

To run the container (I specified port 30000, also fill in AWS credentials that has dynamoDB and secrets manager access): 
docker run -e AWS_ACCESS_KEY_ID=(insert AWS access key) -e AWS_SECRET_ACCESS_KEY=(insert AWS secret access key) -p 30000:8080 cs302-project/workshop:1.0

To run the tests: 
go test ./...