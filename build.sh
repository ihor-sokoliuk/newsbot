#!/bin/bash
project_path="$GOPATH/src/github.com/ihor-sokoliuk/newsbot"
rm -r -f project_path
mkdir -p project_path
cp -r . ${project_path}
cd ${project_path}
echo "Dep ensure..."
go get -u github.com/golang/dep/cmd/dep
dep ensure
./build_news_bot.sh "newsbot" $newsbot_token
./build_news_bot.sh "technewsbot" $technewsbot_token