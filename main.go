package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

func main() {
	portsFlag := flag.String("ports", "1071,2569,5689,6458,7878", "Порты, которые будут слушаться")
	messagesFlag := flag.String(
		"msg",
		"Упс! Попробуй порт 2569,"+
			"опа... кажется это не тот порт... попробуй 5689,"+
			"*карап жатыр* *порт 6458*,"+
			":3 привет\nпривет\nпривет\nпривет\nпривет\nпривет\nпривет\nпривет\nпривет\nпривет\nпривет\nпривет\nпривет\nпривет\n :3 попроюуй порт 7878,"+
			"лаадно хватит с тебя игр твой порт это 8081",
		"Сообщения, которые будут отправлены по данным портам")
	flag.Parse()

	// Проверяем порты
	ports := make([]string, 0)
	for p := range strings.SplitSeq(*portsFlag, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		p = strings.TrimPrefix(p, ":")

		if _, err := strconv.Atoi(p); err != nil {
			log.Fatalf("invalid port %q: %v", p, err)
		}
		ports = append(ports, p)
	}
	if len(ports) == 0 {
		log.Fatal("no ports provided")
	}

	// Проверяем сообщения
	messages := strings.Split(*messagesFlag, ",")
	if len(ports) != len(messages) {
		log.Fatal("messages len isnt equal to ports")
	}

	var wg sync.WaitGroup
	servers := make([]*http.Server, 0, len(ports))

	for i, p := range ports {
		port := p
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Только GET
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintf(w, "method %s not allowed\n", r.Method)
				return
			}
			fmt.Fprintf(w, "%s", messages[i])
		})

		server := &http.Server{
			Addr:    ":" + port,
			Handler: mux,
		}
		servers = append(servers, server)

		wg.Go(func() {
			log.Printf("starting server on :%s", port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("server on :%s error: %v", port, err)
			}
			log.Printf("server on :%s stopped", port)
		})
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutdown")

	var shutWg sync.WaitGroup
	for _, s := range servers {
		shutWg.Go(func() {
			if err := s.Shutdown(context.Background()); err != nil {
				log.Printf("error shutting down server %s: %v", s.Addr, err)
			} else {
				log.Printf("gracefully shut down server %s", s.Addr)
			}
		})
	}

	shutWg.Wait()
	wg.Wait()
	log.Println("all servers stopped")
}
