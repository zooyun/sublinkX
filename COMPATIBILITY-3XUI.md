# 3X-UI / Mihomo 兼容说明

本版本在生成 Clash/Mihomo 订阅时保留 VLESS 链接中的 `encryption`，支持 3X-UI 的 VLESS Encryption 参数，并为 REALITY 输出 `support-x25519mlkem768`。未提供该参数的 REALITY 链接默认开启；对于不支持混合握手的旧服务端，可在原始 VLESS 链接中显式加入 `support-x25519mlkem768=false`。

本版本只处理 VLESS 加密参数、REALITY 的混合握手标志和基本 XHTTP 参数；不宣称支持所有 3X-UI、Xray 或 Mihomo 的高级传输选项。REALITY ML-DSA-65 签名验证与此混合握手标志不同，本版本不提供该验证功能；含有 `pqv` 的节点会明确报错，避免静默丢弃验证参数。

当节点名称与策略组或内置策略名称冲突时，生成的节点会自动加编号，避免形成循环引用。多个节点重名时也会加编号。请检查你的规则、策略组引用和使用节点名的外部自动化。

## 构建与检查

在源码目录运行 `go test ./node` 和 `go build ./...`，或使用 `docker compose up -d --build` 构建并启动。默认绑定服务器本机的 18000 端口，用于经受信任的反向代理或 SSH 隧道访问；先在独立目录中运行，避免覆盖现有实例的数据。不要使用仓库内的旧安装脚本覆盖现有实例：它会从上游下载版本，而不是构建此修改版。镜像包含 CA 证书，以支持 HTTPS 订阅和规则下载。

建议先备份现有实例的 `db`、`template` 和部署配置，再使用独立端口与独立数据目录测试。确认生成的节点 YAML 包含正确的 `encryption` 和 `reality-opts.support-x25519mlkem768` 后，实际测速和访问目标网站；旧节点、其他协议、Xray 内核版本及订阅服务的认证都需要验证。订阅链接、UUID 和密钥应当保密。

Xray v26.9.8 之后的 REALITY 握手、服务端版本限制、TLS 指纹和客户端 Mihomo 内核版本都可能影响连接；本修复不能保证所有组合都能连接。导入的原始 VLESS 链接必须已经包含完整的客户端加密参数；无法从 UUID 或服务端私钥重建遗漏的客户端参数。
