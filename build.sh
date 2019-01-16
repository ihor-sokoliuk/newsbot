#!/bin/bash
echo "Dep ensure..."
go get -u github.com/golang/dep/cmd/dep
dep ensure
./build_news_bot.sh "newsbot" $newsbot_token
./build_news_bot.sh "technewsbot" $technewsbot_token