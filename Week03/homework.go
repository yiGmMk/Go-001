package main

/*任务：
  基于 errgroup 实现一个 http server 的启动和关闭 ，
  以及 linux signal 信号的注册和处理，要保证能够 一个退出，全部注销退出。
*/
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
			fmt.Println("signal,quit")
			quit <- 0
			return errors.New("quit with signal")
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	group.Go(func() error {
		defer func() {
			rec := recover()
			if rec != nil {
				log.Println("panic in group Go func,server shutdown")
			}
		}()
		server = StartNewServer()
		return nil
	})
	group.Go(func() error {
		defer func() {
			rec := recover()
			if rec != nil {
				log.Println("panic in group Go func,server shutdown")
			}
		}()
		fmt.Println("ctx.Err()", ctx.Err())
		select {
		case <-ctx.Done():
			fmt.Println("server shutdown ctx.Done()")
			server.Shutdown(ctx)
			return errors.New("exit with cancle")
		case <-quit:
			fmt.Println("server shutdown")
			err := server.Shutdown(ctx)
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println("server shutdown success")
			return errors.New("exit with signal")
		}
	})

	if err := group.Wait(); err != nil {
		log.Println("exit")
	}
	fmt.Println("ctx.Err()", ctx.Err())
	fmt.Println("normal quit")
}
