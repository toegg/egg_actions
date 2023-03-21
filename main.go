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