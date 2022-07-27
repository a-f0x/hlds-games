rm -r gen
docker run -v $PWD:/defs namely/protoc-all -d .  -l go
sudo chown f0x:f0x -R gen
