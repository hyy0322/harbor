#!/bin/sh

set -e

# The directory /var/lib/registry is within the container, and used to store image in CI testing.
# So for now we need to chown to it to avoid failure in CI.
# if [ -d /var/lib/registry ]; then
#     chown 10000:10000 -R /var/lib/registry
# fi

/home/harbor/install_cert.sh

MEM_LIMIT=$(cat /sys/fs/cgroup/memory/memory.limit_in_bytes)

if [ "$MEM_LIMIT" -eq 9223372036854771712 ]; then
    # 读取主机总内存（字节数）
    HOST_MEM=$(grep MemTotal /proc/meminfo | awk '{print $2 * 1024}')  # MemTotal 单位是 KB，转换为字节
    GOMEMLIMIT=$((HOST_MEM * 9 / 10))
    echo "未设置容器内存限制，使用主机内存的 90%: $GOMEMLIMIT 字节"
else
    GOMEMLIMIT=$((MEM_LIMIT * 9 / 10))
    echo "已设置容器内存限制，使用容器内存的 90%: $GOMEMLIMIT 字节"
fi

export GOMEMLIMIT=$GOMEMLIMIT

exec /usr/bin/registry_DO_NOT_USE_GC serve /etc/registry/config.yml
