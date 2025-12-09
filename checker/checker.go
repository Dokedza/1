package checker

import (
	"1/storage"
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type LinkChecker struct {
	store     storage.Storage
	workers   int
	TaskQueue chan checkTask
	wg        sync.WaitGroup
	cancel    context.CancelFunc
	mu        sync.Mutex
	isRunning bool
}
type checkTask struct {
	setID int
	url   string
}

func NewLinkChecker(store storage.Storage, workers int) *LinkChecker {
	return &LinkChecker{
		store:     store,
		workers:   workers,
		TaskQueue: make(chan checkTask),
	}
}

func (lc *LinkChecker) Start() {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if lc.isRunning {
		log.Println("LinkChecker already running")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	lc.cancel = cancel
	lc.isRunning = true

	for i := 0; i < lc.workers; i++ {
		lc.wg.Add(1)
		go lc.worker(ctx)
	}
	log.Println("Workers :", lc.workers)

}
func (lc *LinkChecker) Stop() {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if !lc.isRunning {
		return
	}

	if lc.cancel != nil {
		lc.cancel()
	}

	close(lc.TaskQueue)
	lc.wg.Wait()
	lc.isRunning = false
}

func (lc *LinkChecker) worker(ctx context.Context) {
	defer lc.wg.Done()
	log.Println("LinkChecker worker started")
	client := http.Client{
		Timeout: time.Second * 10,
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("LinkChecker worker stopped")
			return
		case task, ok := <-lc.TaskQueue:
			log.Printf("Worker: received task - set %d, url %s", task.setID, task.url)
			if !ok {
				log.Println("LinkChecker worker stopped")
				return
			}

			status := lc.checkLink(client, task.url)
			log.Printf("Worker: checking result - %s -> %s", task.url, status)
			lc.store.UpdateLinkStatus(task.setID, task.url, status)
			log.Printf("Worker: checked result - %s -> %s", task.url, status)
		}
	}
}

func (lc *LinkChecker) checkLink(client http.Client, url string) storage.LinkStatus {
	log.Println("checkLink is started")
	fullURL := url
	if !strings.HasPrefix(fullURL, "http://") && !strings.HasPrefix(fullURL, "https://") {
		fullURL = "http://" + fullURL
	}
	resp, err := client.Get(fullURL)
	if err != nil {
		log.Println("checkLink is failed", err)
		return storage.StatusUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		log.Println("checkLink is success")
		return storage.StatusAvailable
	}
	log.Println("checkLink is failed")
	return storage.StatusUnavailable
}

func (lc *LinkChecker) scheduleChecks(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lc.checkAllLinks()
		}
	}
}

func (lc *LinkChecker) checkAllLinks() {
	sets := lc.store.GetAllSets()

	for _, set := range sets {
		for _, link := range set.Links {
			if link.Status == storage.StatusPending {
				select {
				case lc.TaskQueue <- checkTask{setID: set.ID, url: link.URL}:
				default:
					//	очередь заполнена
				}
			}
		}
	}
}

func (lc *LinkChecker) CheckLinksAsync(setId int, urls []string) {
	log.Println("CheckLinksAsync is started")
	for _, url := range urls {
		select {
		case lc.TaskQueue <- checkTask{setID: setId, url: url}:
			log.Println("CheckLinksAsync:", setId, url)
		default:
			log.Println("LinkChecker is busy")
			//	заполнена
		}
	}
}
