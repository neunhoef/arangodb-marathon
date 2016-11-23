all:
	go build arangodb-marathon-framework.go
	strip arangodb-marathon-framework
	docker build -t neunhoef/arangodb-marathon-framework .
	docker push neunhoef/arangodb-marathon-framework
