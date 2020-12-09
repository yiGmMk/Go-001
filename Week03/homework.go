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

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// StartNewServer 启动http server
func StartNewServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("hello，errgroup"))
	})
	server := &http.Server{
		Addr:    "0.0.0.0:2048",
		Handler: mux,
	}
	server.ListenAndServe()
	return server
}

func main() {
	fmt.Println(time.Now().Format("2006-01-02 01:02:59"), "main start")

	group, ctx := errgroup.WithContext(context.Background())
	signalQuit := make(chan os.Signal, 1)

	group.Go(func() error {
		defer func() {
			rec := recover()
			if rec != nil {
				log.Println("panic in group Go func,quit")
			}
		}()
		signal.Notify(signalQuit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
		select {
		case <-ctx.Done():
			fmt.Println("done,quit")
			return errors.New("group err quir")

		case <-signalQuit:
			fmt.Println("signal,quit")
			return errors.New("quit with signal")
		}
	})

	group.Go(func() error {
		defer func() {
			rec := recover()
			if rec != nil {
				log.Println("panic in group Go func,server shutdown")
			}
		}()
		server := StartNewServer()
		select {
		case <-ctx.Done():
			log.Println("server shutdown")
			return server.Shutdown(context.TODO())
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		ctx.Done()
		log.Println("exit")
		close(signalQuit)
	}
	fmt.Println(ctx.Err())
	fmt.Println("normal quit")
}
