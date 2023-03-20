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

##### 第一步：github上开一个空项目

##### 第二步：增加web服务器代码，并推送push到github

服务器代码
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
		w.Write([]byte("hello http HandleFunc"))
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

其中go.mod内容
```go
module github.com/toegg/egg_actions
go 1.16
```

最后把内容推送push到github

##### 第三步：启用项目git actions，直接在github项目中操作
