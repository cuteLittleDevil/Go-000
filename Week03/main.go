package main

import (
	"context"
	"errors"
	"flag"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	var isServerClose bool
	var isCtxClose bool
	flag.BoolVar(&isServerClose, "s", false, "是否模拟服务挂了一个")
	flag.BoolVar(&isCtxClose, "c", false, "是否模拟超时关闭errgroup")
	flag.Parse()

	s1 := &http.Server{
		Addr: ":12345",
	}
	s2 := &http.Server{
		Addr: ":10086",
	}

	closeTask := make(chan struct{})
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	go func() {
		select {
		case <-quit:
			log.Println("quit close")
			close(closeTask)
		}
	}()

	stopTask := func() {
		if err := s1.Shutdown(context.Background()); err != nil {
			log.Println("s1 shutdown err is ", err)
		}
		log.Println("s1 close")
		if err := s2.Shutdown(context.Background()); err != nil {
			log.Println("s2 shutdown err is ", err)
		}
		log.Println("s2 close")
		close(quit)
		<-closeTask
	}

	ctx := context.Background()
	if isCtxClose {
		ctx2, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		ctx = ctx2
		defer cancel()
		log.Println("模拟超时关闭errgroup 等待时间为8s")
	}
	f2 := func() error {
		return s2.ListenAndServe()
	}
	if isServerClose {
		f2 = func() error {
			go s2.ListenAndServe()
			time.Sleep(time.Second * 10)
			return errors.New("测试服务错误")
		}
		log.Println("模拟 s2服务挂了 等待时间为10s")
	}
	g, completeCh := Run(ctx, closeTask, stopTask, func() error {
		return s1.ListenAndServe()
	}, f2)

	if err := g.Wait(); err != nil {
		log.Println("err is ", err)
	}
	<-completeCh
}

func Run(ctx context.Context, closeTask chan struct{}, stopTask func(), tasks ...func() error) (*errgroup.Group, chan struct{}) {
	// 1 开始任务
	g, ctx := errgroup.WithContext(ctx)
	stop := make(chan struct{}, 1)
	var once sync.Once
	for i := 0; i < len(tasks); i++ {
		i := i
		g.Go(func() error {
			err := tasks[i]()
			if err != nil {
				once.Do(func() {
					close(stop)
				})
			}
			return err
		})
	}
	// 2 任务退出

	complete := make(chan struct{})
	go func() {
		select {
		case <-stop: // 某个服务挂了
		case <-ctx.Done(): // errgroup取消
		case <-closeTask: // 其他方式取消
		}
		stopTask()
		close(complete)
	}()
	return g, complete
}
