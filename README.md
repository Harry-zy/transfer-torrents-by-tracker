# transfer-torrents-by-tracker

`transfer-torrents-by-tracker` 是一个基于 Go 的工具，用于在多个 Transmission 服务器之间获取和管理种子。该工具帮助用户根据 Tracker URL、Passkey 过滤种子，并在目录之间复制种子文件，同时允许他们在第二个 Transmission 服务器上上传并设置下载目录。

## Features / 功能

- 连接到多个 Transmission 服务器。
- 根据特定的 Tracker URL 和 Passkey 获取种子。
- 使用哈希字符串去重种子。
- 将过滤后的种子文件复制到指定的目标目录。
- 上传种子并在第二个 Transmission 服务器上设置下载目录。

## Requirements / 需求

- Go 1.18 或更高版本。
- [grequests](https://github.com/levigross/grequests) 库（用于进行 HTTP 请求）。

## Installation / 安装

1. 克隆仓库：
    ```bash
    git clone https://github.com/yourusername/transfer-torrents-by-tracker.git
    ```
2. 导航到项目目录：
    ```bash
    cd transfer-torrents-by-tracker
    ```
3. 安装依赖：
    ```bash
    go get -u github.com/levigross/grequests
    ```

## Usage / 使用

1. 构建项目：
    ```bash
    go build -o transfer-torrents-by-tracker
    ```

2. 运行工具：
    ```bash
    ./transfer-torrents-by-tracker
    ```

3. 按照屏幕提示输入：

    - 第一个 Transmission 服务器地址。
    - 第一个 Transmission 用户名和密码。
    - 第二个 Transmission 服务器地址。
    - 第二个 Transmission 用户名和密码。
    - 源种子目录。
    - 目标种子目录。
    - 用于过滤 Tracker URL 的 Passkey。

### Example / 示例

```bash
请输入第一个 Transmission 服务器地址: http://192.168.1.10:9091
请输入第一个 Transmission 用户名: admin
请输入第一个 Transmission 密码: password123
请输入第二个 Transmission 服务器地址: http://192.168.1.20:9091
请输入第二个 Transmission 用户名: admin2
请输入第二个 Transmission 密码: password456
请输入种子源目录: /path/to/source/torrents
请输入目标种子目录: /path/to/target/torrents
请输入 passkey: your-passkey
>>>>>>> 4e57ec3 (transmission根据tracker转种工具)
