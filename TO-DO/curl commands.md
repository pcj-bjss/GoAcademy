## Curl add command
curl -X POST   -H "Content-Type: application/json" -d '{"Name": "Book taxi", "Due": "27-12-2025"}' http://localhost:8080/create 

## Curl update command
curl -X PATCH -H "Content-Type: application/json" -d '{"Index": 0, "Name": "Book flight for vacation", "Completed": true}' http://localhost:8080/update

## Curl delete command
curl -X DELETE "http://localhost:8080/delete?index=0"

## Curl delete command
curl -X GET "http://localhost:8080/get"