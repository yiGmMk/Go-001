package main

/*任务：
  基于 errgroup 实现一个 http server 的启动和关闭 ，
  以及 linux signal 信号的注册和处理，要保证能够 一个退出，全部注销退出。
*/

// 启动server监听端口会阻塞，需在一个goroutine中单独处理，
// 收到signal后通过channel发送消息给关闭server的chan
// 在Go中都添加一个ctx.Done()用于接收context cancle的结果并退出errgroup的Go函数
import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"errors"

	"golang.org/x/sync/errgroup"
)

// StartNewServer 启动http server
func StartNewServer() *http.Server {
	defer func() {
		rec := recover()
		if rec != nil {
			log.Println("panic in StartNewServer")
		}
	}()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("hello，errgroup"))
		resp.WriteHeader(http.StatusOK)
		time.Sleep(time.Second)
	})
	server := &http.Server{
		Addr:    "0.0.0.0:2248",
		Handler: mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		panic("failed to start server")
	}
	return server
}

func main() {
	fmt.Println(time.Now().Format("2006-01-02 01:02:59"), "main start")

	server := &http.Server{}
	group, ctx := errgroup.WithContext(context.Background())

	signalQuit := make(chan os.Signal)
	quit := make(chan int, 1)

	//接收信号，受到信号发送数据给quit
	group.Go(func() error {
		defer func() {
			rec := recover()
			if rec != nil {
				log.Println("panic in group Go func,quit")
			}
		}()
		signal.Notify(signalQuit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
		select {
		case <-signalQuit:
			log.Println("signal,quit")
			quit <- 0
			return errors.New("quit with signal")
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	//启动 server
	group.Go(func() error {
		defer func() {
			rec := recover()
			if rec != nil {
				log.Println("panic in group Go func in Go to start server")
			}
		}()
		go StartNewServer()
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	})

	//接收quit中数据，收到后关闭server
	group.Go(func() error {
		defer func() {
			rec := recover()
			if rec != nil {
				log.Println("panic in group Go func")
			}
		}()
		fmt.Println("ctx.Err()", ctx.Err())
		select {
		case <-ctx.Done():
			log.Println("server shutdown ctx.Done()")
			server.Shutdown(ctx)
			return errors.New("exit with cancle")
		case <-quit:
			log.Println("server shutdown with msg from quit chan ")
			err := server.Shutdown(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			log.Println("server shutdown success")
			return errors.New("exit with signal")
		}
	})

	if err := group.Wait(); err != nil {
		log.Println("exit")
	}
	log.Println("ctx.Err()", ctx.Err())
	log.Println("normal quit")
}
