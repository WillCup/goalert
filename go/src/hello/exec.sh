. ~/.profile
go build
cp -y apimonitor /usr/local/will/dd/
cd /usr/local/will/dd/
docker build -t willgo:0.1 .
docker tag willgo:0.1 thub.will.com.cn/apimonitor/apimonitor:0.1
docker stop willtest && docker run --name willtest -v `pwd`/config.json:/config.json -d --rm willgo:0.1  /apimonitor
docker logs -f willtest