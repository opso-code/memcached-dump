# memcached-dump

## 说明

`memcached` 脚本工具，支持以下功能

- 查询版本号
- 查询key的总条数
- 列出所有key
- 存储所有数据到文件中（txt）
- 复制一个实例的数据到另一个实例

> 线上环境请谨慎运行

## 运行

```
$ cd build
$ ./memcached-dump version 127.0.0.1:11211
```

```
Usage :
   version <IP:PORT> show the memcached version
   count <IP:PORT>   count the keys
   keys <IP:PORT>    list all keys
   store <IP:PORT>   store data to local file
   dump <IP:PORT> <IP:PORT>  dump all data to another memcached
```

## 注意

建议先用`count`命令运行看看，如果有出现`[use stats cachedump]`的提示，则表示使用的旧的获取key方式，很有可能获取的key不是全部。

由于memcached在`1.4.31`才支持的`lru_crawler metadump`命令，之前的版本只能使用`stats cachedump`命令，有1M的数据大小限制（大概几W个key，看key的长度），所以低版本无法导出完整数据。

还有一种情况是返回了`CLIENT_ERROR lru crawler disabled` 说明memcached启动时`lru crawler`配置没有打开，程序会转而使用旧的方式获取key，也会有1M大小限制。

## 其他

- memcached`1.4.31`新增`lru_crawler metadump`命令，更新日志 [Memcached 1.4.31 Release Notes](https://github.com/memcached/memcached/wiki/ReleaseNotes1431)
- `1.5.1/1.5.2/1.5.3`有个[#issues/667](https://github.com/memcached/memcached/issues/667) ，这个问题导致`lru_crawler metadump all` 命令不返回任何数据。
