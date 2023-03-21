#### Git Actions自动发布部署，非最完善但足够完善和上手的一篇

***

**文章最后附带完整代码链接**

GitHub Actions是一个持续集成和持续交付(CI/CD)平台，允许您自动化构建、测试和部署管道。您可以创建构建和测试存储库中的每个拉取请求的工作流，或者将合并的拉取请求部署到生产中。

* 至于什么是CI/CD?  

CI全称 Continuous Integration，代表持续集成，CD则是 Continuous Delivery和 Continuous Deployment两部分，分别是持续交付和持续部署。CI/CD 是一种软件开发实践，利用自动化的手段来提高软件交付效率，让交付更简单。

以上的概念问题粗暴理解就是，比如我提交push一个代码到git，git actions可以帮助我们实现传输到远程服务器，同时让远程服务器编译，重新启动服务等等。特别是测试阶段，改动代码需要发布并重启服务器，提交git即可自动完成。

* git actions怎么实现呢？ 这里就不讲解太多，尽量以人话过下大概流程。  

通过git actions可以实现on事件监听行为(push,pull事件)，然后触发事件，运行自定义的workflow工作流，workflow工作流执行自定义的job任务，可以把我们想要的行为通过job跑出来。  

相关行为和事件等，都是通过yaml配置文件，入门则了解yaml配置文件中，on, workflow，job，step这几个关键字，以及格式，用法即可，这里就不敞开讲，看[文档](https://docs.github.com/en/actions)

***

#### 下面是一个完整的例子，监听git的push事件，触发传输到远程服务器，平滑关闭服务并重启服务的例子

##### 第一步：github上开一个空项目

##### 第二步：增加web服务器代码，并推送push到github

服务器代码, `main.go`
```go
package  main

import (
	"context"
	"github.com/toegg/egg_actions/test"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)
func main() {
	// 创建一个接收信号的通道,监听signal信息
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("start server")

	//启动web服务器，监听两个API行为，1个测试，1个重启
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Test Print:", test.Test())
		w.Write([]byte("hello http HandleFunc, Result:" + test.Test()))
	})
	http.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		quit <- syscall.SIGINT
		log.Println("Server Reload")
		w.Write([]byte("http HandleFunc"))
	})

	s := &http.Server{
		Addr: ":8881",
	}
	go s.ListenAndServe()

	// 阻塞在此，接收到关闭进程信号，继续往下走
	<-quit
	log.Println("Shutdown Server ...")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(timeoutCtx); err != nil {
		log.Println(err)
	}

}
```

使用module管理，项目目录下执行初始和拉取命令
``shell
go mod init
go get
``

其中go.mod内容如下
```go
module github.com/toegg/egg_actions
go 1.16
```

增加编译的shell脚本, `build.sh`
```go
#!/bin/bash

cd /home/tool/golearn/src/egg_actions

/home/tool/go/go/bin/go build main.go
```

最后把内容推送push到github，必须推送push

##### 第三步：启用项目git actions，直接在github项目中操作

选择Actions，点击build_new_actions

进入自定义yaml配置文件界面，也可以选择提供的模板，右边可选择，不同编程语言等。我们这里用我已经写好的，先贴上去。  
用到了3个`job`，`code`:测试编译go代码，`restart`:平滑关闭服务并重启，`deply`:推送文件模块到远程服务器
```go
# 这里声明git actions的名字
name: Go

# 指定监听事件，这里监听push到git分支main的事件
on:
  push:
    branches: [ "main" ]

# 这里定义工作流，工作流配置对应job任务
jobs:
  code: # 声明job的名字(这个job主要用来介绍，可以用做测试go test流程)
    runs-on: ubuntu-latest   # 使用ubuntu系统镜像运行脚本
    # steps定义job的步骤，具体执行，运行xxx，-字符开头的一块则为一个步骤，可以配置多个步骤
    steps:                   
      - uses: actions/checkout@v2   # 步骤1：下载git仓库中的代码，使用官方提供，使用库用uses关键字
      - name: SetUp Go              # 步骤2：下载go代码
        uses: actions/setup-go@v3   # 使用官方提供
        with:
          go-version: 1.16          # 指定go版本1.16
      - name: Build Go              # 步骤3：编译仓库中的代码
        run: go build -v ./
  restart: # 声明另一个job的名字
    runs-on: ubuntu-latest   
    steps:
      - name: Check my-json-server
        uses: cross-the-world/ssh-pipeline@master   # 使用别人包装好的步骤，执行远程操作的库
        with:
          host: ${{ secrets.REMOTE_HOST }}      # 远程连接的host，引用配置
          user: ${{ secrets.SSH_USERNAME }}     # 远程连接的用户名，引用配置
          key: ${{ secrets.ACCESS_TOKEN }}      # 远程连接的ssh秘钥，引用配置
          port: '22'                            # ssh端口
          connect_timeout: 10s                  # 远程连接的超时时间
          # 下面是执行的脚本内容，找到go服务进程，通过发送信号关闭进程，调用build.sh脚本编译代码，用nohup重启服务
          script: |
            echo "hello egg" 
            (ps aux|grep main|grep -v "grep"|head -n 1|awk '{printf $2}'|xargs kill -15) &&
            (sh /home/tool/golearn/src/test_actions/build.sh) &&
            (nohup /home/tool/golearn/src/test_actions/main > /home/tool/golearn/src/test_actions/error.log 2>&1 &) 
            echo "over"
    needs: deploy   # 依赖关键字，等deploy的执行完了再执行restart
  deploy:
    runs-on: ubuntu-latest   

    steps:  
      - uses: actions/checkout@v2  
      - name: Deploy to Server  # 使用别人包装好的步骤，推送文件模块到远程服务器
        uses: AEnterprise/rsync-deploy@v1.0  
        env:
          DEPLOY_KEY: ${{ secrets.ACCESS_TOKEN }}   #  远程连接的ssh秘钥，引用配置
          ARGS: -avz --delete                       # rsync参数
          SERVER_PORT: '22'                         # ssh端口
          FOLDER: ./                                # 要推送的文件夹，路径相对于代码仓库的根目录
          SERVER_IP: ${{ secrets.REMOTE_HOST }}     # 远程连接的host，引用配置
          USERNAME: ${{ secrets.SSH_USERNAME }}     # 远程连接的用户名，引用配置
          SERVER_DESTINATION: /home/tool/golearn/src/test_actions  # 部署到目标文件的路径
    needs: code # 等code完成再执行deploy

```

粘贴上去，点击**start commit**提交，会自动在项目目录创建`.github/workflows/main.yaml`文件

再选择Actions,会看到我们新增的actions，可以看到我们三个job为线性关系，有前后依赖(因为用到了needs关键字)。  
可以看到在code任务成功了，但是deploy报错停住了，点进去查看具体原因，查看报错原因，原来是由于我们用到了引用配置，但是还没配，执行远程推送服务器失败。

##### 第四步：配置github项目的配置变量，提供引用

进入项目主页，选中settings，点击actions，添加对应的key=》val，分别添加
```go
REMOTE_HOST :远程连接的host
SSH_USERNAME : 远程连接的用户名
ACCESS_TOKEN : 远程连接的ssh秘钥
```

添加完毕后，则如图

##### 第五步：编译go程序可执行文件，手动上传服务器，启动web服务
因为我们为了测试重启，所以先手动上传服务器并启动web服务。


##### 第六步：重新跑actions，测试结果

点击回第三步最后的actions页面，重新跑工作流，测试，选中Re-run all jobs会自动执行。

出现以下结果，测试通过

打开浏览器，访问web服务器链接

##### 第七步：本地修改代码，push推送git，查看是否自动发布，远程传输，重启

修改代码为
```go

```

上github项目查看actions执行结果


打开浏览器，访问web服务器链接

[完整代码链接](https://github.com/toegg/egg_actions)

