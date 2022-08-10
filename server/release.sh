#! /bin/bash
name="saass"
port=16100
branch="main"
configFilePath="config.test.json"
allowMethods=("ls stop gitpull proto dockerremove start logs")

gitpull() {
  echo "-> 正在拉取远程仓库"
  git reset --hard
  git pull origin $branch
}

dockerremove() {
  echo "-> 删除无用镜像"
  docker rm $(docker ps -q -f status=exited)
  docker rmi -f $(docker images | grep '<none>' | awk '{print $3}')
}

start() {
  touch $DIR/conf.json
  echo "-> 正在启动「${name}」服务"
  gitpull
  dockerremove

  echo "-> 正在准备相关资源"
  # 获取npm配置
  DIR=$(cd $(dirname $0) && pwd)
  cp -r ~/.ssh $DIR
  cp -r ~/.gitconfig $DIR

  echo "-> 准备构建Docker"
  docker build \
    -t $name \
    $(cat /etc/hosts | sed 's/^#.*//g' | grep '[0-9][0-9]' | tr "\t" " " | awk '{print "--add-host="$2":"$1 }' | tr '\n' ' ') . \
    -f Dockerfile.multi
  rm -rf $DIR/.ssh
  rm -rf $DIR/.gitconfig

  echo "-> 准备运行Docker"
  remove

  docker run \
    -v $DIR/static:/static \
    -v $DIR/conf.json:/conf.json \
    -v $DIR/$configFilePath:/config.json \
    --name=$name \
    $(cat /etc/hosts | sed 's/^#.*//g' | grep '[0-9][0-9]' | tr "\t" " " | awk '{print "--add-host="$2":"$1 }' | tr '\n' ' ') \
    -p $port:$port \
    --restart=always \
    -d $name
}

ls() {
  start
  logs
}

stop() {
  docker stop $name
}

remove() {
  docker stop $name
  docker rm $name
}

proto() {
  echo "-> 准备编译Protobuf"
  cd ./protos && protoc --go_out=. *.proto

  echo "-> 编译Protobuf成功"
}

logs() {
  docker logs -f $name
}

main() {
  if echo "${allowMethods[@]}" | grep -wq "$1"; then
    "$1"
  else
    echo "Invalid command: $1"
  fi
}

main "$1"
