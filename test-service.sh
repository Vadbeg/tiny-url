curl -X POST http://0.0.0.0:8080/create \
-H "Content-Type: application/json" \
-d '{
  "URL": "http://vtitko.com"
}'

curl -X GET http://0.0.0.0:8080/get_url/9e7dab5d

curl -X GET http://0.0.0.0:8080/get_bindings