#!/bin/bash

readonly RNDSTR="0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
SAID=
SAKEY=
DATAPATH=
SMARTASSISTANT_PORT=
SMARTASSISTANT_GRPC_PORT=
SMARTASSISTANT_HTTP_PORT=
SMARTASSISTANT_HTTPS_PORT=
DOCKER_SOCKET_PATH=
VERSION="1.5.0"

#######################################
# Writes error message
# Arguments:
#   Anything
# Outputs:
#   Writes error message to stderr
#######################################
err() {
  echo "[$(date +'%Y-%m-%dT%H:%M:%S%z')]: $*" >&2
}

#######################################
# Environment check
# Arguments:
#   None
#######################################
environment_check() {

  openssl version 2>&1 > /dev/null
  if [ $? != 0 ]; then
    err "please install openssl!"
    exit 1  
  fi

  docker --version 2>&1 > /dev/null
  if [ $? != 0 ]; then
    err "please install docker!"
    exit 1  
  fi

  docker-compose --version 2>&1 > /dev/null
  if [ $? != 0 ]; then
    err "please install docker-compose!"
    exit 1  
  fi

  docker ps 2>&1 > /dev/null
  if [ $? != 0 ]; then
    err "please start dockerd!"
    exit 1  
  fi
}

#######################################
# Random Number [$1, $2)
# Globals:
#   RANDOM
# Arguments:
#   $1 Range Start
#   $2 Range End
# Outputs:
#   A Random Number
#######################################
rnd2() {
  if [ -z "${RANDOM}" ] ; then
    seed=$(tr -cd 0-9 </dev/urandom | head -c 8)
  else
    seed=${RANDOM}
  fi

  rnd_num=$(echo ${seed} $1 $2|awk '{srand($1);printf "%d",rand()*10000%($3-$2)+$2}')
  echo ${rnd_num}
}

#######################################
# Random String 
# Globals:
#   RNDSTR
# Arguments:
#   $1 Ramdom String Length
# Outputs:
#   Random String
#######################################
rand_string() {
  ret=""
  local length=${#RNDSTR}

  for i in $(seq $1); do
    index=$(rnd2 0 ${length})
    # use command
    # ret=${ret}"${RNDSTR:${index}:1}"
    ret=${ret}$(echo ${RNDSTR} | awk '{print substr($0, "'"${index}"'", 1)}')
  done

  echo $ret
}


#######################################
# Build Path, like golang path.Join() 
# Globals:
#   DATAPATH
# Arguments:
#   Anything
# Outputs:
#   filepath
#######################################
build_path() {
  # use command
  # new_path=${DATAPATH}
  # tail=${DATAPATH##*/}
  # if [ -z ${tail} ]; then
  #   new_path=${DATAPATH%%/*}
  # fi
  new_path=${DATAPATH}
  tail=$(echo ${new_path} | awk '{print substr($0,length())}')
  if [ "${tail}" == "/" ]; then
    new_path=$(echo ${new_path} | awk '{print substr($0, 0,length()-1)}')
  fi

  for p in "$@"; do
    new_path=${new_path}"/"${p}
  done
  echo ${new_path}
}


#######################################
# Initialization docker-compose config file
# Globals:
#   SMARTASSISTANT_PORT
#   SMARTASSISTANT_GRPC_PORT
#   DOCKER_SOCKET_PATH
#   DATAPATH
# Arguments:
#   None
# Outputs:
#   Writes content to docker-compose config file
#######################################
init_docker_compose(){
cat > $(build_path "docker-compose.yaml") <<EOF
version: "3.3"

services:
  zt-vue:
    image: zhitingtech/zt-nginx:1.5.0
    ports:
      - ${SMARTASSISTANT_HTTP_PORT}:${SMARTASSISTANT_HTTP_PORT}
      - ${SMARTASSISTANT_HTTPS_PORT}:${SMARTASSISTANT_HTTPS_PORT}
    volumes:
      - ./zt-nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./zt-nginx/test-scene.conf:/etc/nginx/conf.d/test-scene.conf
      - ./certs:/etc/nginx/certs
    depends_on:
      - smartassistant
  etcd:
    image: zhitingtech/etcd:3.5.1
    ports:
      - 2379:2379
      - 2380:2380
  fluentd:
    image: fluent/fluentd:v1.13
    environment:
      - FLUENTD_CONF=fluentd.conf
    ports:
      - "24224:24224"
      - "24224:24224/udp"
    volumes:
      - ./config/fluentd.conf:/fluentd/etc/fluentd.conf
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "10"
  smartassistant:
    image: zhitingtech/smartassistant:${VERSION}
    ports:
      - "${SMARTASSISTANT_PORT}:${SMARTASSISTANT_PORT}"
      - "${SMARTASSISTANT_GRPC_PORT}:${SMARTASSISTANT_GRPC_PORT}"
      - "54321:54321/udp"
    logging:
      driver: "fluentd"
      options:
        fluentd-address: "localhost:24224"
        tag: smartassistant.main
        fluentd-async-connect: "true"
    volumes:
      - ${DOCKER_SOCKET_PATH}:/var/run/docker.sock
      - type: bind
        source: ${DATAPATH}
        target: ${DATAPATH}
    depends_on:
      - etcd
      - fluentd

  supervisor:
    image: zhitingtech/supervisor:${VERSION}
    volumes:
      - ${DOCKER_SOCKET_PATH}:/var/run/docker.sock

volumes:
  db:

EOF

if [ $? -ne 0 ]; then
  err "Init docker-compose.yaml error"
  exit 1
else
  echo "Init docker-compose.yaml Complete!"
fi
}

#######################################
# Initialization smartassistant config file
# Globals:
#   SMARTASSISTANT_PORT
#   SMARTASSISTANT_GRPC_PORT
#   DATAPATH
#   SAID
#   SAKEY
# Arguments:
#   None
# Outputs:
#   Writes content to smartassistant config file
#######################################
init_smartassistant_config(){
cat > $(build_path "config" "smartassistant.yaml") <<EOF
debug: true
smartcloud:
  domain: "scgz.zhitingtech.com"
  tls: true

smartassistant:
  id: "${SAID}"
  key: "${SAKEY}"
  host: 0.0.0.0
  port: ${SMARTASSISTANT_PORT}
  grpc_port: ${SMARTASSISTANT_GRPC_PORT}
  runtime_path: "${DATAPATH}"
  host_runtime_path: "${DATAPATH}"
  database:
    driver: sqlite

docker:
  username: ""
  password: ""

datatunnel:
  export_services:
    http: "zt-vue:${SMARTASSISTANT_HTTP_PORT}"
    https: "zt-vue:${SMARTASSISTANT_HTTPS_PORT}"
EOF

if [ $? -ne 0 ]; then
  err "Init smartassistant.yaml error"
  exit 1
else
  echo "Init smartassistant.yaml Complete!"
fi
}

#######################################
# Initialization fluentd config file
# Globals:
#   SAID
#   SAKEY
# Arguments:
#   None
# Outputs:
#   Writes content to fluentd config file
#######################################
init_fluentd_config(){
cat > $(build_path "config" "fluentd.conf") <<EOF
<source>
  @type forward
</source>

<filter smartassistant.*>
  @type parser
  key_name log
  reserve_time true
  <parse>
    @type json
    time_key time
    time_type string
    time_format %Y-%m-%dT%H:%M:%S
    keep_time_key true
  </parse>
</filter>

<match smartassistant.*>
  @type copy
  <store>
    @type stdout
  </store>
  <store>
    @type http

    endpoint http://127.0.0.1:8082/api/log_replay
    open_timeout 2
    http_method post

    <format>
      @type json
    </format>
    <buffer>
      flush_interval 10s
    </buffer>
    <auth>
      method basic
      username ${SAID}
      password ${SAKEY}
    </auth>
  </store>
</match>
EOF

if [ $? -ne 0 ]; then
  err "Init fluentd.conf error"
  exit 1
else
  echo "Init fluentd.conf Complete!"
fi
}


#######################################
# Initialization nginx config file
# Globals:
#   SMARTASSISTANT_HTTP_PORT
#   SMARTASSISTANT_HTTPS_PORT
# Arguments:
#   None
# Outputs:
#   Writes content to nginx config file
#######################################
init_nginx_config(){
cat > $(build_path "zt-nginx" "nginx.conf") <<EOF
user  nginx;
worker_processes  auto;

error_log  /var/log/nginx/error.log notice;
pid        /var/run/nginx.pid;


events {
    worker_connections  1024;
}


http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    log_format  main  '\$remote_addr - \$remote_user [\$time_local] "\$request" '
                      '\$status \$body_bytes_sent "\$http_referer" '
                      '"\$http_user_agent" "\$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;

    client_max_body_size 200m;
    sendfile        on;
    #tcp_nopush     on;

    keepalive_timeout  65;

    #gzip  on;

    include /etc/nginx/conf.d/*.conf;
}
EOF

if [ $? -ne 0 ]; then
  err "Init nginx.conf error"
  exit 1
else
  echo "Init nginx.conf Complete!"
fi


cat > $(build_path "zt-nginx" "test-scene.conf") <<EOF
server {
    listen       ${SMARTASSISTANT_HTTP_PORT};
    listen       ${SMARTASSISTANT_HTTPS_PORT} ssl;
    server_name  sa.zhitingtech.com ;
    ssl_certificate /etc/nginx/certs/sa.zhitingtech.com.pem ;
    ssl_certificate_key /etc/nginx/certs/sa.zhitingtech.com.key;
    ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE:ECDH:AES:HIGH:!NULL:!aNULL:!MD5:!ADH:!RC4:!DH:!DHE;

    access_log  /var/log/nginx/test-scene.log  main;
    error_page   500 502 503 504  /50x.html;
    client_max_body_size 10M;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }

    #access_log  /var/log/nginx/host.access.log  main;

    location / {
        alias /home/test-scene/;
        index  index.html index.htm;
    }

    location /ws {
        proxy_pass http://smartassistant:${SMARTASSISTANT_PORT}/ws;
        proxy_http_version 1.1;
        proxy_read_timeout 360s;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    location /api {
        proxy_set_header Host  \$http_host;
        proxy_set_header X-Scheme \$scheme;

        proxy_pass   http://smartassistant:${SMARTASSISTANT_PORT};
    }
}
EOF

if [ $? -ne 0 ]; then
  err "Init test-scene.conf error"
  exit 1
else
  echo "Init test-scene.conf Complete!"
fi

}

#######################################
# Generate rsa key and ca certificate
# Arguments:
#   None
#######################################
init_https (){
  key_path=$(build_path "certs" "sa.zhitingtech.com.key")
  openssl genrsa -out ${key_path} 2048
  if [ $? -ne 0 ]; then
    err "Generate RSA key error"
    exit 1 
  else
    echo "Generate RSA key Complete!"
  fi

  csr_path=$(build_path "certs" "sa.zhitingtech.com.csr")
  openssl req -new -key ${key_path} -out ${csr_path} -subj "/CN=sa.zhitingtech.com"
  if [ $? -ne 0 ]; then
    err "Generate csr error"
    exit 1 
  else
    echo "Generate csr Complete!"
  fi

  ca_path=$(build_path "certs" "sa.zhitingtech.com.pem")
  openssl x509 -req -in ${csr_path} -out ${ca_path} -signkey ${key_path} -days 3650 \
    -extfile <(printf "subjectAltName=DNS:sa.zhitingtech.com\nbasicConstraints=CA:true\nsubjectKeyIdentifier=hash\nauthorityKeyIdentifier=keyid:always,issuer")
  if [ $? -ne 0 ]; then
    err "Generate ca certificate error"
    exit 1 
  else
    echo "Generate ca certificate Complete!"
  fi
}

#######################################
# Initialization directory
# Arguments:
#   None
#######################################
init_dir() {

  dir=$(build_path "zt-nginx")
  rm -rf ${dir}
  mkdir -p ${dir}
  if [ $? -ne 0 ]; then
    err "Create directory ${dir} error"
    exit 1
  else
    echo "Create directory ${dir} Complete!"
  fi

  dir=$(build_path "certs")
  rm -rf ${dir}
  mkdir -p ${dir}
  if [ $? -ne 0 ]; then
    err "Create directory ${dir} error"
    exit 1
  else
    echo "Create directory ${dir} Complete!"
  fi

  dir=$(build_path "config")
  rm -rf ${dir}
  mkdir -p ${dir}
  if [ $? -ne 0 ]; then
    err "Create directory ${dir} error"
    exit 1
  else
    echo "Create directory ${dir} Complete!"
  fi

  dir=$(build_path "data" "smartassistant")
  rm -rf ${dir}
  mkdir -p ${dir}
  if [ $? -ne 0 ]; then
    err "Create directory ${dir} error"
    exit 1
  else
    echo "Create directory ${dir} Complete!"
  fi
}

check_port() {
  expr $1 + 0 &>/dev/null
  if [ $? -eq 0 ]; then
    if [ $1 -gt 65535 ]; then
      err "Invaild port!"
      exit 1
    elif [ $1 -le 0 ]; then
      err "Invaild port!"
      exit 1
    fi
  else 
    err "Invaild port!"
    exit 1
  fi
}

#######################################
# Input Value
# Globals:
#   SMARTASSISTANT_PORT
#   SMARTASSISTANT_GRPC_PORT
#   DOCKER_SOCKET_PATH
#   DATAPATH
#   SAID
#   SAKEY
# Arguments:
#   None
#######################################
init_val() {
  read -p "Enter SAID(default random):"                                SAID
  if [ -z ${SAID} ]; then
    SAID=$(rand_string 10)-sa
  fi

  read -p "Enter SAKEY(default random):"                               SAKEY
  if [ -z ${SAKEY} ]; then
    SAKEY=$(rand_string 15)
  fi

  read -p "Enter Data Full Path(default /mnt/data/zt-smartassistant):" DATAPATH
  if [ -z ${DATAPATH} ]; then 
    DATAPATH="/mnt/data/zt-smartassistant"
  elif [ ${DATAPATH:0:1} != "/" ]; then
    err "Please enter the full path!"
    exit 1
  fi

  read -p "Enter Smartassistant Port(default 37965):"                  SMARTASSISTANT_PORT
  if [ -z "${SMARTASSISTANT_PORT}" ]; then
    SMARTASSISTANT_PORT=37965
  else 
    check_port ${SMARTASSISTANT_PORT}
  fi

  read -p "Enter Smartassistant Grpc Port(default 9234):"              SMARTASSISTANT_GRPC_PORT
  if [ -z "${SMARTASSISTANT_GRPC_PORT}" ]; then
    SMARTASSISTANT_GRPC_PORT=9234
  else 
    check_port ${SMARTASSISTANT_GRPC_PORT}
  fi

  read -p "Enter Smartassistant Http Port(default 9020):"              SMARTASSISTANT_HTTP_PORT
  if [ -z "${SMARTASSISTANT_HTTP_PORT}" ]; then
    SMARTASSISTANT_HTTP_PORT=9020
  else 
    check_port ${SMARTASSISTANT_HTTP_PORT}
  fi

  read -p "Enter Smartassistant Https Port(default 9030):"             SMARTASSISTANT_HTTPS_PORT
  if [ -z "${SMARTASSISTANT_HTTPS_PORT}" ]; then
    SMARTASSISTANT_HTTPS_PORT=9030
  else 
    check_port ${SMARTASSISTANT_HTTPS_PORT}
  fi

  read -p "Enter Docker Socket Path(default /var/run/docker.sock):"    DOCKER_SOCKET_PATH
  if [ -z ${DOCKER_SOCKET_PATH} ]; then 
    DOCKER_SOCKET_PATH="/var/run/docker.sock"
  elif [ ${DOCKER_SOCKET_PATH:0:1} != "/" ]; then
    err "Please enter the full path!"
    exit 1
  fi

  echo 
  echo "Config Table:"
  echo "SAID:                        ${SAID}"
  echo "SAKEY:                       ${SAKEY}"
  echo "Data Path:                   ${DATAPATH}"
  echo "Smartassistant Port:         ${SMARTASSISTANT_PORT}"
  echo "Smartassistant Grpc Port:    ${SMARTASSISTANT_GRPC_PORT}"
  echo "Smartassistant Http Port:    ${SMARTASSISTANT_HTTP_PORT}"
  echo "Smartassistant Https Port:   ${SMARTASSISTANT_HTTPS_PORT}"
  echo "Docker Socket Path:          ${DOCKER_SOCKET_PATH}"

  read -p "Check Your Config(Y/n):" start
  if [ "${start}" == "Y" ] || [ "${start}" == "y" ]; then
    echo "Start Init Config!"
  else 
    echo "Stop Init!"
    exit 0
  fi  
}

main() {
  environment_check
  init_val
  init_dir
  init_smartassistant_config
  init_fluentd_config
  init_docker_compose
  init_https
  init_nginx_config

  echo "Config Init Finish! Ready To Start!"

  cd ${DATAPATH} && docker-compose up -d
  
  if [ $? -ne 0 ]; then 
    err "Start error"
    exit 1
  else
    echo "Smartassistant Start Complete!!!"
  fi
}

main $*