# SSH 端口映射

这个小工具可以实现远程端口批量映射，解决实际开发中需要**SSH TUNNEL**才能访问的网络环境

## 使用方法

1. **配置文件`forwarding.yaml`**

    ```yaml
    # 日志级别 DEBUG(0), INFO(1), WARN(2), ERROR(3)
    logging:
      level: 1
    # 映射配置  
    ssh_servers:
      - host: <SSH 地址>
        port: <SSH 端口>
        username: <SSH 用户名>
        password: <SSH 密码>
        forwardings:
          - label: <自定义标签 1>
            local_port: <本地端口1>
            remote_host: <远程地址1>
            remote_port: <远程端口1>
          - label: <自定义标签 2>
            local_port: <本地端口2>
            remote_host: <远程地址2>
            remote_port: <远程端口2>
    ```
    配置文件加载顺序:

       1. 命令执行目录
       2. 脚本程序所在目录
       3. 用户家目录`~/.ssh-forwarding/`


2. **编译**

    ```bash
    go build 
    ```

3. **运行**
    ```
    ./ssh-forwarding
    ```