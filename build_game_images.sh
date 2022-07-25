DOCKER_BUILDKIT=1 docker build --output type=local,dest=out .
echo "Build cs-nosteam image"
docker build -f ./Dockerfile_cs -t af0x/hlds-agent:cs-nosteam .
echo "Build csdm-nosteam image"
docker build -f ./Dockerfile_csdm -t af0x/hlds-agent:csdm-nosteam .
echo "Build hl-nosteam image"
docker build -f ./Dockerfile_hl -t af0x/hlds-agent:hl-nosteam .
