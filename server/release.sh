#! /bin/bash
name="saass"
runName="$name-run"
port=16100
DIR=$(cd $(dirname $0) && pwd)
branch="main"
configFilePath="config.pro.json"
allowMethods=("runLogs run unzip backup ls stop remove gitpull proto dockerremove start logs")

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
  echo "" >>./conf.json
  echo "-> 正在启动「${name}」服务"
  # gitpull
  dockerremove

  echo "-> 正在准备相关资源"
  # 获取npm配置
  cp -r ~/.ssh $DIR
  cp -r ~/.gitconfig $DIR
  git config --global url."git@github.com:".insteadOf "https://github.com/"

  echo "-> 准备构建Docker"
  docker build \
    -t $name \
    --network host \
    $(cat /etc/hosts | sed 's/^#.*//g' | grep '[0-9][0-9]' | tr "\t" " " | awk '{print "--add-host="$2":"$1 }' | tr '\n' ' ') . \
    -f Dockerfile.multi
  rm -rf $DIR/.ssh
  rm -rf $DIR/.gitconfig

  echo "-> 准备运行Docker"
  stop

  docker run \
    -v $DIR/static:/static \
    -v $DIR/web:/web \
    -v $DIR/conf.json:/conf.json \
    -v $DIR/$configFilePath:/config.json \
    --name=$name \
    $(cat /etc/hosts | sed 's/^#.*//g' | grep '[0-9][0-9]' | tr "\t" " " | awk '{print "--add-host="$2":"$1 }' | tr '\n' ' ') \
    -p $port:$port \
    --restart=always \
    -d $name

  echo "-> 整理文件资源"
  docker cp $name:/saass $DIR/saass
  stop

  ./ssh.sh run

  rm -rf $DIR/saass
}

run() {
  echo "-> 正在启动「${runName}」服务"
  dockerremove

  echo "-> 准备构建Docker"
  docker build \
    -t \
    $runName \
    --network host \
    . \
    -f Dockerfile.run.multi

  echo "-> 准备运行Docker"
  stop
  docker run \
    -v $DIR/static:/static \
    -v $DIR/web:/web \
    -v $DIR/conf.json:/conf.json \
    -v $DIR/$configFilePath:/config.json \
    --name=$runName \
    $(cat /etc/hosts | sed 's/^#.*//g' | grep '[0-9][0-9]' | tr "\t" " " | awk '{print "--add-host="$2":"$1 }' | tr '\n' ' ') \
    -p $port:$port \
    --restart=always \
    -d $runName
}

ls() {
  start
  logs
}

stop() {
  docker stop $name
  docker rm $name
  docker stop $runName
  docker rm $runName
}

backup() {
  # backupTime=$(date +'%Y-%m-%d_%T')
  # zip -q -r ./saass_$backupTime.zip ./static
  tar cvzf /home/project/static/saass_static.tgz -C $DIR/static .

  # unzip -d ./ build_2023-07-04_21:11:13.zip
}

unzip() {
  # unzip -d ./ /home/static/saass_static.zip
  mkdir -p $DIR/static
  tar -zxvf $DIR/saass_static.tgz \
    -C $DIR/static
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

runLogs() {
  docker logs -f $runName
}

main() {
  if echo "${allowMethods[@]}" | grep -wq "$1"; then
    "$1"
  else
    echo "Invalid command: $1"
  fi
}

main "$1"
