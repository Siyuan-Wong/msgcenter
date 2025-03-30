#!/bin/bash
set -euo pipefail

# 默认配置
REDIS_PASSWORD="Wsymt1203."
CONSUL_ADVERTISE_ADDR="196.168.1.43"
PULSAR_ENABLE=true
CONSUL_ENABLE=true
REDIS_ENABLE=true
PULSAR_MANAGER_ENABLE=true
CLEANUP=false
REDIS_CONF_URL="https://raw.githubusercontent.com/redis/redis/unstable/redis.conf"
PULSAR_MANAGER_USER="admin"
PULSAR_MANAGER_PASSWORD="apachepulsar"
PULSAR_MANAGER_EMAIL="admin@pulsar-manager.local"

# 容器名称（使用时间戳避免冲突）
PULSAR_CONTAINER="pulsar-$(date +%s)"
CONSUL_CONTAINER="consul-$(date +%s)"
REDIS_CONTAINER="redis-$(date +%s)"
PULSAR_MANAGER_CONTAINER="pulsar-manager-$(date +%s)"

# 帮助信息
usage() {
  cat <<EOF
Usage: $0 [-r redis_password] [-c consul_advertise_address] [-p] [-o] [-d] [-m] [-x] [-h]
Options:
  -r <password>    Redis密码 (必需)
  -c <address>     Consul外网地址 (必需)
  -p              禁用Pulsar (可选)
  -o              禁用Consul (可选)
  -d              禁用Redis (可选)
  -m              禁用Pulsar Manager (可选)
  -x              启动前清理已有资源 (可选)
  -h              显示帮助信息
EOF
  exit 1
}

# 清理资源函数
cleanup() {
  echo "正在清理资源..."

  for container in $PULSAR_CONTAINER $CONSUL_CONTAINER $REDIS_CONTAINER $PULSAR_MANAGER_CONTAINER; do
    if docker ps -a --format '{{.Names}}' | grep -q "^${container}\$"; then
      echo "移除容器 $container"
      docker rm -f $container >/dev/null 2>&1 || true
    fi
  done

  rm -f /tmp/redis.conf
}

# 检查端口冲突
check_port() {
  local port=$1 service=$2
  if netstat -tuln | grep -q ":$port "; then
    echo "错误: 端口 $port 已被占用 ($service)"
    exit 1
  fi
}

# 下载Redis配置文件
download_redis_conf() {
  echo "下载Redis官方配置文件..."
  if ! curl -sSL $REDIS_CONF_URL -o /tmp/redis.conf; then
    echo "备用方案：从Redis镜像提取默认配置"
    docker run --rm redis cat /usr/local/etc/redis/redis.conf > /tmp/redis.conf || {
      echo "获取Redis配置失败"
      exit 1
    }
  fi
}

# 设置Pulsar Manager密码（不使用jq）
init_pulsar_manager() {
  echo "初始化Pulsar Manager..."

  # 等待服务就绪
  local attempt=0 max_attempts=30
  while ! curl -s http://localhost:7750/pulsar-manager/csrf-token >/dev/null; do
    attempt=$((attempt + 1))
    if [ $attempt -gt $max_attempts ]; then
      echo "错误: Pulsar Manager启动超时"
      exit 1
    fi
    sleep 2
  done

  CSRF_TOKEN=$(curl -s http://localhost:7750/pulsar-manager/csrf-token | jq -r '.token')
  if [ -z "$CSRF_TOKEN" ]; then
    echo "错误: 无法获取CSRF Token"
    exit 1
  fi

  # 创建管理员用户
  if ! curl -sSf \
    -H "X-XSRF-TOKEN: $CSRF_TOKEN" \
    -H "Cookie: XSRF-TOKEN=$CSRF_TOKEN;" \
    -H "Content-Type: application/json" \
    -X PUT http://localhost:7750/pulsar-manager/users/superuser \
    -d '{
      "name": "'"$PULSAR_MANAGER_USER"'",
      "password": "'"$PULSAR_MANAGER_PASSWORD"'",
      "description": "Administrator",
      "email": "'"$PULSAR_MANAGER_EMAIL"'"
    }'; then
    echo "警告: 用户创建失败（可能已存在）"
  fi

  echo "Pulsar Manager管理员账号:"
  echo "用户名: $PULSAR_MANAGER_USER"
  echo "密码: $PULSAR_MANAGER_PASSWORD"
}


# 解析参数
while getopts "r:c:podmxh" opt; do
  case $opt in
    r) REDIS_PASSWORD="$OPTARG" ;;
    c) CONSUL_ADVERTISE_ADDR="$OPTARG" ;;
    p) PULSAR_ENABLE=false ;;
    o) CONSUL_ENABLE=false ;;
    d) REDIS_ENABLE=false ;;
    m) PULSAR_MANAGER_ENABLE=false ;;
    x) CLEANUP=true ;;
    h) usage ;;
    *) usage ;;
  esac
done

# 验证必需参数
if [ -z "$REDIS_PASSWORD" ] && [ "$REDIS_ENABLE" = true ]; then
  echo "错误: Redis密码必须通过 -r 参数指定"
  usage
fi

if [ -z "$CONSUL_ADVERTISE_ADDR" ] && [ "$CONSUL_ENABLE" = true ]; then
  echo "错误: Consul外网地址必须通过 -c 参数指定"
  usage
fi

# 执行清理
if [ "$CLEANUP" = true ]; then
  cleanup
fi

# 检查端口冲突
echo "检查端口冲突..."
[ "$PULSAR_ENABLE" = true ] && check_port 6650 "Pulsar" && check_port 8080 "Pulsar"
[ "$CONSUL_ENABLE" = true ] && check_port 8300 "Consul" && check_port 8500 "Consul"
[ "$REDIS_ENABLE" = true ] && check_port 6379 "Redis"
[ "$PULSAR_MANAGER_ENABLE" = true ] && check_port 9527 "Pulsar Manager" && check_port 7750 "Pulsar Manager API"

# 打印配置
echo "=== 服务启动配置 ==="
echo "Redis密码: [${REDIS_PASSWORD:-未设置}]"
echo "Consul地址: [${CONSUL_ADVERTISE_ADDR:-未设置}]"
echo "Pulsar启用: $PULSAR_ENABLE"
echo "Consul启用: $CONSUL_ENABLE"
echo "Redis启用: $REDIS_ENABLE"
echo "Pulsar Manager启用: $PULSAR_MANAGER_ENABLE"
[ "$PULSAR_MANAGER_ENABLE" = true ] && echo "Pulsar Manager管理员: $PULSAR_MANAGER_USER / $PULSAR_MANAGER_PASSWORD"
echo "=================="

# 拉取镜像
echo "拉取Docker镜像..."
if [ "$PULSAR_ENABLE" = true ]; then
  docker pull apachepulsar/pulsar:latest
fi
if [ "$CONSUL_ENABLE" = true ]; then
  docker pull hashicorp/consul:latest
fi
if [ "$REDIS_ENABLE" = true ]; then
  docker pull redis:latest
fi
if [ "$PULSAR_MANAGER_ENABLE" = true ]; then
  docker pull apachepulsar/pulsar-manager:latest
fi

# 创建数据卷
echo "初始化数据卷..."
if [ "$PULSAR_ENABLE" = true ]; then
  docker volume create pulsar-data >/dev/null || echo "pulsar-data卷已存在"
  docker volume create pulsar-conf >/dev/null || echo "pulsar-conf卷已存在"
fi
if [ "$CONSUL_ENABLE" = true ]; then
  docker volume create consul-data >/dev/null || echo "consul-data卷已存在"
fi
if [ "$REDIS_ENABLE" = true ]; then
  docker volume create redis-data >/dev/null || echo "redis-data卷已存在"
  docker volume create redis-config >/dev/null || echo "redis-config卷已存在"
  download_redis_conf
  sed -i.bak \
    -e 's/^bind 127.0.0.1 -::1/bind 0.0.0.0/' \
    -e 's/^protected-mode yes/protected-mode no/' \
    -e "s/^# requirepass .*/requirepass $REDIS_PASSWORD/" \
    /tmp/redis.conf
  docker run --rm -v redis-config:/data -v /tmp/redis.conf:/tmp/redis.conf alpine \
    cp /tmp/redis.conf /data/redis.conf
  rm -f /tmp/redis.conf /tmp/redis.conf.bak
fi
if [ "$PULSAR_MANAGER_ENABLE" = true ]; then
  docker volume create pulsar-manager-data >/dev/null || echo "pulsar-manager-data卷已存在"
fi

# 启动服务
echo "启动服务..."

## Pulsar
if [ "$PULSAR_ENABLE" = true ]; then
  echo "启动Pulsar ($PULSAR_CONTAINER)..."
  docker run -d --name $PULSAR_CONTAINER \
    -p 6650:6650 \
    -p 8080:8080 \
    -v pulsar-data:/pulsar/data \
    -v pulsar-conf:/pulsar/conf \
    apachepulsar/pulsar bin/pulsar standalone

  echo "Pulsar服务URL:"
  echo "  - Broker: pulsar://localhost:6650"
  echo "  - Admin: http://localhost:8080"
fi

## Consul
if [ "$CONSUL_ENABLE" = true ]; then
  echo "启动Consul ($CONSUL_CONTAINER)..."
  docker run -d --name $CONSUL_CONTAINER \
    -p 8300:8300 -p 8500:8500 -p 8600:8600 \
    -v consul-data:/consul/data \
    hashicorp/consul agent -server -ui -node=n1 -bootstrap-expect=1 -client=0.0.0.0 -advertise=$CONSUL_ADVERTISE_ADDR

  echo "Consul服务URL:"
  echo "  - UI: http://localhost:8500"
fi

## Redis
if [ "$REDIS_ENABLE" = true ]; then
  echo "启动Redis ($REDIS_CONTAINER)..."
  docker run -d --name $REDIS_CONTAINER \
    -p 6379:6379 \
    -v redis-data:/data \
    -v redis-config:/usr/local/etc/redis \
    redis redis-server /usr/local/etc/redis/redis.conf

  echo "Redis服务URL:"
  echo "  - redis://:${REDIS_PASSWORD}@localhost:6379"
fi

## Pulsar Manager
if [ "$PULSAR_MANAGER_ENABLE" = true ]; then
  echo "启动Pulsar Manager ($PULSAR_MANAGER_CONTAINER)..."
  docker run -d --name $PULSAR_MANAGER_CONTAINER \
    -p 9527:9527 \
    -p 7750:7750 \
    -v pulsar-manager-data:/data \
    -e SPRING_CONFIGURATION_FILE=/pulsar-manager/pulsar-manager/application.properties \
    apachepulsar/pulsar-manager:latest

  init_pulsar_manager

  echo "Pulsar Manager服务URL:"
  echo "  - UI: http://localhost:9527"
  echo "  - API: http://localhost:7750"
fi

# 输出汇总信息
cat <<EOF

=== 服务启动完成 ===
容器名称:
$([ "$PULSAR_ENABLE" = true ] && echo "Pulsar: $PULSAR_CONTAINER")
$([ "$CONSUL_ENABLE" = true ] && echo "Consul: $CONSUL_CONTAINER")
$([ "$REDIS_ENABLE" = true ] && echo "Redis: $REDIS_CONTAINER")
$([ "$PULSAR_MANAGER_ENABLE" = true ] && echo "Pulsar Manager: $PULSAR_MANAGER_CONTAINER")

访问信息:
$([ "$PULSAR_MANAGER_ENABLE" = true ] && echo "Pulsar Manager: http://localhost:9527 (用户: $PULSAR_MANAGER_USER 密码: $PULSAR_MANAGER_PASSWORD)")
EOF
