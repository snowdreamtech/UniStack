# Quickstart: 客户端对软件仓库的集成

## 测试验证流程

1. 启动本地静态服务器作为远程 Registry
   ```bash
   python3 -m http.server 8080 --directory ansible/roles/apps/
   ```
   
2. 在另一个终端测试下载更新
   ```bash
   # 测试同步逻辑
   UNISTACK_REGISTRY_URL="http://localhost:8080/packages.db.zst" unistack update
   
   # 验证 packages.db 是否已解压到缓存目录
   ls ~/.unistack/cache/packages.db
   
   # 测试下载包
   unistack download hello
   
   # 验证文件是否落盘
   ls ~/.unistack/cache/hello-1.0.0.tar.gz
   ```
