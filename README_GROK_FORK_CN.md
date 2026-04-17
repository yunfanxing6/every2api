# every2api

这是基于 `sub2api v0.1.113` 整理出的 `every2api` 工作目录，目标是：

- 保留 `Sub2API` 原有的用户、API Key、余额、订阅、审计能力
- 保留并增强独立 `grok` / `qwen` / `any2api` 平台路由
- 让 `every2api` 对接 `any2api` 作为统一上游网关
- 让 `any2api` 再集成 `grok2api` 和 `qwen2API`
- 对外继续提供文本、文生图、文生视频和媒体文件代理

## 当前状态

- 仓库名：`every2api`
- 基线版本：`v0.1.113`

## 已融合能力

- 独立平台：`grok`
- Grok 默认模型：
  - `grok-4.20-beta`
  - `grok-imagine-1.0-fast`
  - `grok-imagine-1.0-edit`
  - `grok-imagine-1.0-video`
- OpenAI 兼容路由：
  - `/v1/chat/completions`
  - `/v1/messages`
  - `/v1/models`
- 媒体路由：
  - `/v1/images/generations`
  - `/v1/images/edits`
  - `/v1/videos`
  - `/v1/video/extend`
  - `/v1/files/image/*`
  - `/v1/files/video/*`
- Grok 分组价格配置：
  - 文本输入/输出
  - 图片 1K/2K
  - 视频 5s/10s/15s
  - 高清倍率

## 部署思路

推荐保持三层结构：

1. `every2api` 对外提供用户 API、用户系统、Key、分组和计费
2. `any2api` 作为统一 Grok/Qwen 上游网关
3. `grok2api` 与 `qwen2API` 只作为 `any2api` 内部集成组件

推荐配置：

- `every2api` 域名：例如 `https://sub.example.com`
- `any2api` 域名：例如 `https://api.example.com`
- 在 `every2api` 后台创建：
  - 平台：`any2api`
  - 账号类型：`apikey`
  - `base_url`：`https://api.example.com/v1`
  - `api_key`：你的 any2api API Key

其中：

- `every2api` 负责对用户暴露 `grok` / `qwen` 模型和计费能力
- `any2api` 负责把 Grok 请求下发给 `grok2api`，把 Qwen 请求下发给 `qwen2API`

## 线上验证状态

当前这套融合版已验证：

- 文本：可用
- 文生图：可用
- 文生视频：可用
- 媒体文件代理：可用

已知未完全打通：

- `grok-imagine-1.0-edit`

这个问题目前更像 `grok2api` 上游编辑能力适配问题，不是 `every2api` 集成层的问题。

## 开源建议

如果你要把这套代码直接开源，建议：

1. 新建仓库，例如：`every2api`
2. 以这个目录为仓库根目录推送
3. 在仓库首页说明这不是官方 `sub2api`，而是面向 Grok / Qwen 的外层网关分支
4. 明确标注当前已知限制：`images/edits`
5. 单独维护升级策略：以后以官方 `sub2api` tag 为基线继续前移

## 升级方式

### 已经在用这个 fork 的用户

这个 fork 的在线更新默认检查：

- `yunfanxing6/every2api`

如果你以后换了自己的仓库名，可以通过环境变量覆盖：

- `SUB2API_RELEASE_REPO=owner/repo`

### 还在使用原版 Sub2API 的用户

可以使用一键迁移脚本：

```bash
curl -sSL https://raw.githubusercontent.com/yunfanxing6/every2api/main/deploy/upgrade-to-grok-fork.sh | bash
```

这个脚本会：

- 备份当前部署文件
- 自动识别现有 compose 文件
- 先停止旧服务再切换
- 保留 `.env` 和数据目录
- 克隆本 fork 到本地
- 生成 `docker-compose.grok.yml`
- 本地构建并启动 Grok 融合版
- 生成回滚脚本

## 当前线上实例

- 站点：`https://sub.openaiapi.icu`
- Grok 上游：`https://grok.openaiapi.icu`
