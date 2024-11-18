# httpcheck

- 简介
批量探测目标是否存活的轮子。为什么要造轮子? httpx很好用，但是不知道是bug还是使用参数问题，已经漏了好几次目标了。。。


- 使用
  
```bash
git clone https://github.com/Al0neme/httpcheck.git
cd httpcheck
go build .
./httpcheck -f t.txt -t 5
```

- 结果

![{2BEEC156-6E89-43A2-866F-51E9977E7110}.png](https://al0neme-staticfile.oss-cn-hangzhou.aliyuncs.com/static/202411182317685.png)
