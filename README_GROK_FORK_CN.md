# Sub2API Grok Fork

这个目录是基于 `sub2api v0.1.113` 整理出的 Grok 融合版工作目录，目标是：

- 保留 `Sub2API` 原有的用户、API Key、余额、订阅、审计能力
- 新增独立 `grok` 平台标签
- 让 `sub2api` 直接对接 `grok2api` 作为 Grok 上游
- 支持文本、文生图、文生视频和媒体文件代理

## 当前目录

- 工作目录：`/home/xingyunfan/sub2api-grok`
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

推荐保持两层结构：

1. `sub2api-grok` 对外提供用户 API
2. `grok2api` 只作为内网/专用上游

推荐配置：

- `sub2api` 域名：例如 `https://sub.example.com`
- `grok2api` 域名：例如 `https://grok.example.com`
- 在 `sub2api` 后台创建：
  - 平台：`grok`
  - 账号类型：`apikey`
  - `base_url`：`https://grok.example.com/v1`
  - `api_key`：你的 grok2api API Key

## 线上验证状态

当前这套融合版已验证：

- 文本：可用
- 文生图：可用
- 文生视频：可用
- 媒体文件代理：可用

已知未完全打通：

- `grok-imagine-1.0-edit`

这个问题目前更像 `grok2api` 上游编辑能力适配问题，不是 `sub2api` 集成层的问题。

## 开源建议

如果你要把这套代码直接开源，建议：

1. 新建仓库，例如：`sub2api-grok`
2. 以这个目录为仓库根目录推送
3. 在仓库首页说明这不是官方 `sub2api`，而是 Grok 融合分支
4. 明确标注当前已知限制：`images/edits`
5. 单独维护升级策略：以后以官方 `sub2api` tag 为基线继续前移

## 升级方式

### 已经在用这个 fork 的用户

这个 fork 的在线更新默认检查：

- `yunfanxing6/sub2api-grok`

如果你以后换了自己的仓库名，可以通过环境变量覆盖：

- `SUB2API_RELEASE_REPO=owner/repo`

### 还在使用原版 Sub2API 的用户

可以使用一键迁移脚本：

```bash
curl -sSL https://raw.githubusercontent.com/yunfanxing6/sub2api-grok/main/deploy/upgrade-to-grok-fork.sh | bash
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
