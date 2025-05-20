# Linux Process Exporter

这是一个用Go语言编写的Prometheus exporter，用于收集Linux系统中所有进程的信息，包括PID、进程名称、CPU使用率和内存使用率。

## 功能特点

- 收集所有进程的基本信息（PID和进程名称）
- 监控每个进程的CPU使用率
- 监控每个进程的内存使用率
- 支持TLS加密和基本认证
- 通过HTTP接口暴露Prometheus格式的metrics
- 支持Docker部署

## 安装

### 从源码安装

```bash
# 克隆仓库
git clone [repository-url]
cd linux-process-exporter

# 安装依赖
go mod download

# 编译
go build
```

### Docker部署

1. 使用Docker Compose（推荐）：
```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

2. 使用Docker命令：
```bash
# 构建镜像
docker build -t linux-process-exporter .

# 运行容器
docker run -d \
  --name process-exporter \
  --pid=host \
  --privileged \
  -v /proc:/host/proc:ro \
  -v /sys:/host/sys:ro \
  -p 2112:2112 \
  linux-process-exporter
```

## 配置

### 命令行参数

- `--version`: 显示版本信息
- `--help`: 显示帮助信息
- `--web.config.file`: TLS和基本认证配置文件路径
- `--web.listen-address`: Web服务监听地址（默认：:2112）

### TLS和认证配置

创建`web-config.yml`文件（参考`web-config.yml.example`）：

```yaml
# TLS配置
tls_server_config:
  cert_file: server.crt
  key_file: server.key

# 基本认证配置
basic_auth_users:
  admin: $2y$10$kHSXPKX.SqaDe3zJ7HN5h.SqaDe3zJ7HN5h # 密码为admin
```

## Metrics说明

- `process_info`：进程基本信息，标签包含pid和name
  - 示例：`process_info{pid="1",name="systemd"} 1`
- `process_cpu_usage`：进程CPU使用率（百分比）
  - 示例：`process_cpu_usage{pid="1",name="systemd"} 0.5`
- `process_memory_usage`：进程内存使用率（百分比）
  - 示例：`process_memory_usage{pid="1",name="systemd"} 1.2`

## Prometheus配置

在Prometheus的配置文件中添加以下job：

```yaml
scrape_configs:
  - job_name: 'process-exporter'
    static_configs:
      - targets: ['localhost:2112']
    # 如果启用了基本认证，添加以下配置
    basic_auth:
      username: admin
      password: admin
    # 如果启用了TLS，添加以下配置
    scheme: https
    tls_config:
      insecure_skip_verify: true
```

## 示例查询

以下是一些常用的PromQL查询示例：

1. 查询CPU使用率最高的5个进程：
```
topk(5, process_cpu_usage)
```

2. 查询内存使用率超过50%的进程：
```
process_memory_usage > 50
```

3. 查询特定进程名称的资源使用情况：
```
process_cpu_usage{name="nginx"}
process_memory_usage{name="nginx"}
```



## Grafana展示示例

* 进程资源使用情况展示

![image-20250518160905004](https://lsky-img.hzbb.top/EAFluSPqdFTVhvgii4ENaXGjGntQVKdn/2025/05/18/682995a60d15b.png)

* 数据关联及标签名称配置

![image-20250518161052206](https://lsky-img.hzbb.top/EAFluSPqdFTVhvgii4ENaXGjGntQVKdn/2025/05/18/6829960e4dab7.png)

