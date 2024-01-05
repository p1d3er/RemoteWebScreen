# RemoteWebScreen

## 项目概述

本项目是一个远程控制应用，使用 Golang 开发，允许用户通过 Web 界面远程控制和屏幕监控其他计算机。主要功能包括屏幕共享、鼠标和键盘控制以及键盘记录。

## 工具使用

注：启动工具时，关闭一下防火墙。此工具基于正向连接，所以会在被控端启动端口。

```
Windows server 2003及之前版本：
netsh firewall set opmode disable  #关闭  
netsh firewall set opmode enable   #开启
Windows server 2003之后版本：
netsh advfirewall set allprofiles state off  #关闭    
netsh advfirewall set allprofiles state on   #开启
```

```
RemoteWebScreen.exe [端口号]      #默认443
```

```
https://IP:端口号/:端口号         #屏幕控制
https://IP:端口号/:端口号log      #键盘记录
```

### 屏幕控制

注：非管理员运行时启动任务管理器，鼠标键盘控制会被禁止。

访问`https://IP:端口号/:端口号`。访问需要安装证书

![image-20231124095233832](image-20231124095233832.png)

以上三处分别为，`切换到扩展屏`、`鼠标键盘控制`、`画质修改`。

**退出杀软**

可以直接通过模拟鼠标退出`火绒`。其他杀软未测试，针对`360`因为360有HOOK鼠标键盘操作所以不建议使用鼠标键盘控制，因为会失效。

![image-20231124101731491](image-20231124101731491.png)

### 键盘记录

注：项目结束时请清理生成的文件

访问`https://IP:端口号/:端口号log	`

当有键盘记录时会生成记录文件到以下目录

```
%tmp%/screen_log/templog.tmp								#注:键盘记录
%tmp%/screen/2006_01_02_15_04_05_04.png					#注:截屏记录
```

![image-20231124101333601](image-20231124101333602.png)

通过上图可以记录到输入的账号密码，同时当用户打开密码本复制密码时，也能获取`Ctrl+c/v`，同时当用户进行复制和粘贴操作时会截一张图。

![image-20231124101600198](image-20231124101600198.png)

## 安装证书

双击安装证书

```
RemoteWebScreen.p12  #密码:RemoteWebScreen
RemoteWebScreen.exe  sha256:ce1f60c29574e0d6d23adfbe90b5bf97119d784c33894985aae1a8cafeba3291
RemoteWebScreen.exe  md5:290d91b5f8b512738a62e0b2d5f0b0fa
```
注：小技巧，缩放浏览器也可以调节画面清晰度。欢迎issues

**仅供技术研究使用，请勿用于非法用途，否则后果作者概不负责**
