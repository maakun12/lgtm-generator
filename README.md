# lgtm-generator

## SetUp
```
docker-compose run app sh
go install
```

## Create LambdaFunction zip
In Container
```
go build main.go
zip function.zip main
```
