# memcached-dump

## 说明

`memcached` 脚本工具，支持以下功能

- stats 查询stats信息
- count 查询key的总条数
- keys 列出所有key
- dump 导出所有数据到文件中（txt）
- transfer 复制一个实例的数据到另一个实例

## :rotating_light: 注意

由于memcached在`1.4.31`才支持的`lru_crawler metadump`命令，之前的版本只能使用`stats cachedump`命令，有1M的数据大小限制（大概几W个key，看key的长度），所以低版本可能无法导出完整数据。

还有一种情况是返回了`CLIENT_ERROR lru crawler disabled` 说明memcached启动时`lru crawler`配置没有打开，程序会转而使用旧的方式获取key，也会有1M大小限制。

`1.5.1/1.5.2/1.5.3`有个[#issues/667](https://github.com/memcached/memcached/issues/667) ，这个问题会导致`lru_crawler metadump all` 命令不返回任何数据。

> 总之，如果你的memcached实例版本是大于`1.4.31`，且不为`1.5.1/1.5.2/1.5.3`，可以尝试使用这个工具导出所有数据，否则可能数据导出不全。

## 运行

```bash
# 查询memcahed信息
$ ./memcached-dump stats 127.0.0.1:11211
# 导出所有数据到本地文件
$ ./memcached-dump dump 127.0.0.1:11211
# 帮助
$ ./memcached-dump -h
```
