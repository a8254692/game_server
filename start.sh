#! /bin/sh

ENV_NAME=$1
ERRMSG="local|dev|prod"
if [ ${ENV_NAME} = "" ]; then
    echo $ERRMSG
    exit
fi

mkdir -p startLog/

echo "================start shell==================="


killall -9 DBServer

nohup ./DBServer ${ENV_NAME} > startLog/DBServerNohup.log &

ps -ef | grep DBServer | grep  -v grep

echo "========================================="


killall -9 OtherServer

nohup ./OtherServer ${ENV_NAME} > startLog/OtherServer.log &

ps -ef | grep OtherServer | grep  -v grep

echo "========================================="


killall -9 GameServer

nohup ./GameServer ${ENV_NAME} > startLog/GameServerNohup.log &

ps -ef | grep GameServer | grep  -v grep

echo "========================================="


killall -9 GateWsServer

nohup ./GateWsServer ${ENV_NAME} > startLog/GateWsServerNohup.log &

ps -ef | grep GateWsServer | grep  -v grep

echo "========================================="

killall -9 WebServer

nohup ./WebServer > startLog/WebServerNohup.log &

ps -ef | grep WebServer | grep  -v grep

echo "================end shell==================="