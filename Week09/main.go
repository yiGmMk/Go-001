package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	group, _ := errgroup.WithContext(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

	group.Go(func() error {
		defer func() {
			rec := recover()
			if rec != nil {
				log.Println(rec)
			}
		}()
		select {
		case <-signalChan:
			log.Println("signal,quit")
			return errors.New("quit with signal")
		case <-ctx.Done():
			return errors.WithStack(ctx.Err())
		}
	})

	group.Go(func() error {
		defer func() {
			rec := recover()
			if rec != nil {
				log.Println(rec)
			}
		}()
		listen, err := net.Listen("tcp", "127.0.0.1:19999")
		if err != nil {
			return errors.WithStack(err)
		}

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				conn, err := listen.Accept()
				if err != nil {
					return errors.WithStack(err)
				}

				ch := make(chan []byte)
				defer conn.Close()

				group.Go(func() error {
					rd := bufio.NewReader(conn)
					for {
						select {
						case <-ctx.Done():
							return errors.WithStack(ctx.Err())
						default:
							line, _, err := rd.ReadLine()
							if err != nil {
								close(ch)
								return err
							}
							ch <- line
						}
					}
					return errors.New("quit read")
				})
				group.Go(func() error {
					wr := bufio.NewWriter(conn)
					for {
						select {
						case <-ctx.Done():
							return errors.WithStack(ctx.Err())
						default:
							line, ok := <-ch
							if !ok {
								return nil
							}
							if len(line) <= 0 {
								wr.Write([]byte("-----\n"))
								continue
							}
							wr.WriteString("Replay ")
							wr.Write(line)
							wr.WriteString("\n")
							wr.Flush()
						}
					}
					return errors.New("quit write")
				})
			}
		}

		return errors.New("accept error ")
	})
	if err := group.Wait(); err != nil {
		log.Println(err)
	}
}
