# build deko 
echo "==================================="
echo "           Building Deko           "
echo "==================================="
echo ""
echo ""
echo ""
echo ""
echo ""


CGO_ENABLED=0 GOOS=linux go build -o "deko" 


# run docker compose 

echo "==================================="
echo "       Building Docker Image       "
echo "==================================="
echo ""
echo ""
echo ""
echo ""
echo ""


docker compose -f docker-compose.yml build 

echo "==================================="
echo "      Running Docker Container     "
echo "==================================="
echo ""
echo ""
echo ""
echo ""

docker compose -f docker-compose.yml up -d
