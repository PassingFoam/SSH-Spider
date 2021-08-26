# SSH_Spider by go



> 目标：内网从一个端点自动爬取证书并喷射


## 起因

1. 日常在内网渗透中为了方便进行秘钥传递，自动化减少手工
2. 看到有大佬用python写了一个版本，不够优雅，而且全部都是手工提取，最多只能爬取七层内网，有些拉胯，且是固定只能爬取root下id_rsa内的私钥，使用场景十分
3. 于是自己整了一个版本

## 设计思路
> 大致思路，具体细节可以看代码
1. 输入一个私钥和ip及用户名
2. 登录之后，自动爬起root目录下以及home下的用户中所有的私钥，通过ssh私钥解析工具判断是否有效，有效则加入私钥库中
3. 同理获取所有know_hosts，用这些私钥去碰撞，如果可以成功登录就返回，记录成为下一层可用ip
4. 将收集到的这一层ip使用递归重复2-3，直到没有know_hosts，或者没有收集到私钥的一层则停止返回
5. 将收集的全部私钥保留

## 使用方法
* Usage of ./SSH_Spider:
    * -ip string
        * input ip your want to start
    * -pkey string
        * input the path of privatekey
    * -username string
        * input username (default "root")


## TODO
[+] 完成回溯法,减少输出层数

[+] 做多线程，目前还是单线程