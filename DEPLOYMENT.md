# CI/CD 部署指南

## 概述

本仓库使用 GitHub Actions 实现 CI/CD 流程，基于 **GitHub Flow** 分支策略：

- **`main`** - 唯一长期分支，始终处于可部署状态
- **`feature/*`** - 功能分支，开发完成后通过 PR 合并到 main

### 流程图

```
┌────────────────────────────────────────────────────────────────┐
│                        代码提交 (Push)                         │
└─────────────────────────┬──────────────────────────────────────┘
                          │
                          ▼
┌────────────────────────────────────────────────────────────────┐
│                      CI 阶段 (自动触发)                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  1. Build & Test                                        │   │
│  │     - go build                                          │   │
│  │     - go test                                           │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────┬──────────────────────────────────────┘
                          │ ✓ 通过
                          ▼
┌───────────────────────────────────────────────────────────────┐
│                   CD - 测试环境 (自动部署)                    │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │  2. Build Docker Image                                  │  │
│  │     - 构建 Docker 镜像                                  │  │
│  │     - 推送到 ghcr.io                                    │  │
│  │                                                         │  │
│  │  3. Deploy to Staging                                   │  │
│  │     - SSH 登录服务器                                    │  │
│  │     - 拉取最新镜像                                      │  │
│  │     - 重启容器                                          │  │
│  └─────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
```

### 触发条件

| 事件 | 触发条件 | 执行操作 |
|------|----------|----------|
| Push 到 main | `git push` | CI + 部署到 staging |
| Pull Request | PR 到 main | 仅 CI (build & test) |
| 创建版本 Tag | `git tag v1.0.0 && git push --tags` | 部署到 production（需启用） |

---

## 生产环境部署

当前 workflow 中生产环境部署配置默认是注释状态。如需启用，请按以下步骤操作：

### 启用生产部署

1. **取消注释 workflow.yml 中的生产部署配置**（第 105-131 行）

2. **添加生产环境 Secrets**

   进入仓库 → **Settings** → **Secrets and variables** → **Actions** → **New repository secret**

   | Secret 名称 | 值 | 说明 |
   |------------|-----|------|
   | `SSH_HOST_PROD` | `1.2.3.4` | 生产服务器公网 IP |
   | `SSH_USER_PROD` | `root` 或 `ubuntu` | 生产服务器用户名 |
   | `SSH_PRIVATE_KEY_PROD` | 私钥文件内容 | 生产环境私钥 |

   > **建议**：生产环境使用独立的 SSH 密钥对，与测试环境隔离

3. **（可选）配置环境审批人**

   进入 **Settings** → **Environments** → `production`：
   - 添加 **Required reviewers**
   - 添加审批人后，每次生产部署前需要人工批准

### 部署到生产环境

```bash
# 1. 测试环境验证通过后，创建版本标签
git tag v1.0.0

# 2. 推送标签触发生产部署
git push origin v1.0.0

# 或一次性完成
git tag v1.0.0 && git push origin v1.0.0
```

### 生产环境部署脚本

启用后，生产部署会执行：

```bash
# 1. 登录容器镜像仓库
echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_ACTOR --password-stdin

# 2. 拉取带版本号的镜像
docker pull ghcr.io/ddddincn/ci-cd-example:<version-tag>

# 3. 停止并删除旧容器
docker stop ci-cd-example || true
docker rm ci-cd-example || true

# 4. 启动新容器
docker run -d --name ci-cd-example -p 8080:8080 --restart=always ghcr.io/ddddincn/ci-cd-example:<version-tag>

# 5. 清理旧镜像
docker image prune -f
```

---

## 第一步：准备服务器

### 1.1 安装 Docker

```bash
# 一键安装 Docker
curl -fsSL https://get.docker.com | sh

# 验证安装
docker --version

# 将当前用户加入 docker 组（避免每次都用 sudo）
sudo usermod -aG docker $USER

# 刷新组设置（或重新登录）
newgrp docker

# 验证
docker ps
```

### 1.2 配置防火墙

```bash
# 开放 8080 端口（应用端口）
sudo ufw allow 8080/tcp

# 如果使用云服务器，还需在云控制台开放相应端口
```

---

## 第二步：生成 SSH 密钥

### 2.1 在本地生成密钥

```bash
# 生成专用密钥（不要使用你的个人 SSH 密钥）
ssh-keygen -t ed25519 -f ~/.github_actions_deploy_key -C "github-actions-deploy"
```

执行后会生成两个文件：
- `~/.github_actions_deploy_key` - **私钥**（保密，不上传）
- `~/.github_actions_deploy_key.pub` - **公钥**（上传到服务器）

### 2.2 将公钥复制到服务器

```bash
# 复制公钥到服务器
ssh-copy-id -i ~/.github_actions_deploy_key.pub user@your-server-ip

# 测试连接
ssh -i ~/.github_actions_deploy_key user@your-server-ip
```

如果 `ssh-copy-id` 不可用，手动复制：

```bash
# 查看公钥内容
cat ~/.github_actions_deploy_key.pub

# 在服务器上执行
mkdir -p ~/.ssh
echo "上面复制的公钥内容" >> ~/.ssh/authorized_keys
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys
```

---

## 第三步：在 GitHub 配置 Secrets

进入仓库 → **Settings** → **Secrets and variables** → **Actions** → **New repository secret**

### 添加以下 Secrets：

| Secret 名称 | 值 | 说明 |
|------------|-----|------|
| `SSH_HOST_DEV` | `1.2.3.4` | 服务器公网 IP |
| `SSH_USER_DEV` | `root` 或 `ubuntu` | 服务器用户名 |
| `SSH_PRIVATE_KEY_DEV` | 复制 `~/.github_actions_deploy_key` 文件内容 | 私钥（包含 BEGIN/END 行） |

> **注意**：私钥内容必须完整，包括：
> ```
> -----BEGIN OPENSSH PRIVATE KEY-----
> ...（多行内容）...
> -----END OPENSSH PRIVATE KEY-----
> ```

---

## 第四步：测试部署

### 4.1 推送代码测试

```bash
# 推送到 main，触发 staging 部署
git add .
git commit -m "test: trigger deployment"
git push origin main
```

### 4.2 查看部署状态

在 GitHub 仓库页面：
1. 点击 **Actions** 标签
2. 选择正在运行的 workflow
3. 查看 `deploy-staging` job 的输出

### 4.3 验证部署

```bash
# SSH 登录服务器查看
ssh user@your-server-ip

# 查看容器状态
docker ps

# 查看应用日志
docker logs -f ci-cd-example

# 测试应用
curl http://localhost:8080
```

---

## 故障排查

### 问题 1: SSH 连接失败

```
ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey]
```

**原因**：SSH 密钥配置问题

**解决**：
```bash
# 在本地测试 SSH 连接
ssh -i ~/.github_actions_deploy_key user@your-server-ip

# 如果失败，检查：
# 1. 服务器防火墙是否开放 22 端口
# 2. SSH 公钥是否正确添加到 ~/.ssh/authorized_keys
# 3. GitHub Secret 中的私钥是否正确（包含 BEGIN/END 行）
```

### 问题 2: Docker 权限不足

```
permission denied while trying to connect to the Docker daemon socket
```

**原因**：用户不在 docker 组

**解决**：
```bash
# 将用户加入 docker 组
sudo usermod -aG docker $USER
newgrp docker  # 或重新登录
```

### 问题 3: 端口被占用

```
Bind for 127.0.0.1:8080 failed: port is already allocated
```

**原因**：8080 端口已被其他进程占用

**解决**：
```bash
# 查找占用 8080 端口的进程
sudo lsof -i :8080
# 或
sudo netstat -tlnp | grep 8080

# 杀死占用端口的进程
sudo kill -9 <PID>

# 或者清理所有已停止的容器
docker container prune -f
```

### 问题 4: 镜像名称大小写

```
invalid reference format: repository name must be lowercase
```

**原因**：Docker 镜像名称必须全部小写

**解决**：当前 workflow 已配置使用小写名称 `ci-cd-example`，无需额外操作。

---

## 常见问题

### Q: 可以使用同一台服务器作为 staging 和 production 吗？

A: 可以，但建议区分不同的容器名称和端口：
- staging: `ci-cd-example`，端口 `8080`
- production: `ci-cd-example-prod`，端口 `8081`

### Q: 如何回滚？

A: 使用之前的 commit 重新部署：
```bash
# 回滚到上一个 commit
git revert HEAD
git push origin main
```

### Q: 如何查看历史部署记录？

A: GitHub → Actions → 选择 workflow run → 查看部署日志

---

## 附录：Workflow 配置说明

### 环境变量

| 变量名 | 值 | 说明 |
|--------|-----|------|
| `REGISTRY` | `ghcr.io` | 容器镜像仓库 |
| `RepositoryName` | `ci-cd-example` | 小写仓库名 |

### 部署脚本内容

每次部署会执行以下命令：

```bash
# 1. 登录容器镜像仓库
echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_ACTOR --password-stdin

# 2. 拉取最新镜像
docker pull ghcr.io/ddddincn/ci-cd-example:latest

# 3. 停止并删除旧容器
docker stop ci-cd-example || true
docker rm ci-cd-example || true

# 4. 启动新容器
docker run -d --name ci-cd-example -p 8080:8080 --restart=always ghcr.io/ddddincn/ci-cd-example:latest

# 5. 清理旧镜像
docker image prune -f
```
