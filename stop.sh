#! /bin/sh

echo "================start shell==================="

killall -9 OtherServer

ps -ef | grep OtherServer | grep  -v grep

echo "========================================="


killall -9 DBServer

ps -ef | grep DBServer | grep  -v grep

echo "========================================="


killall -9 GameServer

ps -ef | grep GameServer | grep  -v grep

echo "========================================="


killall -9 GateWsServer

ps -ef | grep GateWsServer | grep  -v grep

echo "========================================="

killall -9 WebServer

ps -ef | grep WebServer | grep  -v grep

echo "================end shell==================="