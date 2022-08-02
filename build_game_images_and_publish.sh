./build_game_images.sh
echo "Publish cs-nosteam image"
docker push af0x/hlds-agent:cs-nosteam
echo "Publish csdm-nosteam image"
docker push af0x/hlds-agent:csdm-nosteam
echo "Publish hl-nosteam image"
docker push af0x/hlds-agent:hl-nosteam