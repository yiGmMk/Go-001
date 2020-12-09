package main

/*任务：
  基于 errgroup 实现一个 http server 的启动和关闭 ，
  以及 linux signal 信号的注册和处理，要保证能够 一个退出，全部注销退出。
*/
import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func main() {
	fmt.Println(time.Now().Format("2006-01-02 01:02:59"), "main start")

	group, ctx := errgroup.WithContext(context.Background())
	quit := make(chan os.Signal, 0)
	charChan := make(chan int64, 0)
	group.Go(func() error {
		select {
		case <-ctx.Done():

			fmt.Println("done,quit")
			return errors.New("group err quir")
		case <-quit:

			fmt.Println("signal,quit")
			return nil
		case <-charChan:
			ctx.Done()
			return errors.New("char")
		}
	})

	group.Go(func() error {
		for range time.Tick(time.Second * 1) {
			fmt.Println(time.Now().Format("2006-01-02 01:02:59"))

			if time.Now().Second() == 55 {
				quit <- syscall.SIGINT
			}
		}
		return nil
	})

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	if err := group.Wait(); err != nil {
		ctx.Done()
		log.Println("exit")
	}

	fmt.Println("normal quit")
}
func main11() {
	m, err := MD5All(context.Background(), ".")
	if err != nil {
		log.Fatal(err)
	}

	for k, sum := range m {
		fmt.Printf("%s:\t%x\n", k, sum)
	}
}

type result struct {
	path string
	sum  [md5.Size]byte
}

// MD5All reads all the files in the file tree rooted at root and returns a map
// from file path to the MD5 sum of the file's contents. If the directory walk
// fails or any read operation fails, MD5All returns an error.
func MD5All(ctx context.Context, root string) (map[string][md5.Size]byte, error) {
	// ctx is canceled when g.Wait() returns. When this version of MD5All returns
	// - even in case of error! - we know that all of the goroutines have finished
	// and the memory they were using can be garbage-collected.
	g, ctx := errgroup.WithContext(ctx)
	paths := make(chan string)

	g.Go(func() error {
		defer close(paths)
		return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			select {
			case paths <- path:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	})

	// Start a fixed number of goroutines to read and digest files.
	c := make(chan result)
	const numDigesters = 20
	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for path := range paths {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				select {
				case c <- result{path, md5.Sum(data)}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}
	go func() {
		g.Wait()
		close(c)
	}()

	m := make(map[string][md5.Size]byte)
	for r := range c {
		m[r.path] = r.sum
	}
	// Check whether any of the goroutines failed. Since g is accumulating the
	// errors, we don't need to send them (or check for them) in the individual
	// results sent on the channel.
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return m, nil
}
