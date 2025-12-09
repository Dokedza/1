package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type LinkStatus string

const (
	StatusAvailable   LinkStatus = "available"
	StatusUnavailable LinkStatus = "unavailable"
	StatusPending     LinkStatus = "pending"
)

type LinkResult struct {
	URL    string     `json:"url"`
	Status LinkStatus `json:"status"`
}
type LinkSet struct {
	ID        int          `json:"id"`
	Links     []LinkResult `json:"links"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type Storage interface {
	SaveLinks(url []string) (int, error)
	GetLinkSet(id int) (*LinkSet, bool)
	GetLinkSets(ids []int) ([]LinkSet, error)
	UpdateLinkStatus(id int, url string, status LinkStatus) error
	GetAllSets() []LinkSet
	Backup() error
	Restore() error
}

type fileStorage struct {
	mu       sync.Mutex
	sets     map[int]*LinkSet
	nextID   int
	filePath string
}

func NewFileStorage(filePath string) *fileStorage {
	return &fileStorage{
		sets:     make(map[int]*LinkSet),
		nextID:   1,
		filePath: filePath,
	}
}

func (s *fileStorage) SaveLinks(urls []string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID
	s.nextID++

	results := make([]LinkResult, len(urls))
	for i, url := range urls {
		results[i] = LinkResult{
			URL:    normalizeURL(url),
			Status: StatusPending,
		}
	}
	set := &LinkSet{
		ID:        id,
		Links:     results,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.sets[id] = set
	return id, nil
}

func (s *fileStorage) GetLinkSets(ids []int) ([]LinkSet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]LinkSet, 0, len(ids))
	for _, id := range ids {
		set, ok := s.sets[id]
		if ok {
			result = append(result, *set)
		}
	}
	return result, nil
}

func (s *fileStorage) GetLinkSet(id int) (*LinkSet, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	set, ok := s.sets[id]
	if !ok {
		return nil, false
	}
	return set, true
}

func (s *fileStorage) UpdateLinkStatus(id int, url string, status LinkStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	set, ok := s.sets[id]
	if !ok {
		return fmt.Errorf("cannot update link status for id %d", id)
	}
	normUrl := normalizeURL(url)
	log.Println("Updating status by url", normUrl, status)
	for i := range set.Links {
		log.Println("Updating link status", i, status)
		if set.Links[i].URL == normUrl {
			set.Links[i].Status = status
			set.UpdatedAt = time.Now()
			log.Println("Updating is successful")
			return nil
		}
	}
	return fmt.Errorf("url %s is not found in set %d", url, id)
}

func (s *fileStorage) GetAllSets() []LinkSet {
	s.mu.Lock()
	defer s.mu.Unlock()

	results := make([]LinkSet, 0, len(s.sets))
	for _, set := range s.sets {
		results = append(results, *set)
	}
	return results
}

func (s *fileStorage) Backup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := struct {
		Sets   map[int]*LinkSet `json:"sets"`
		NextID int              `json:"next_id"`
	}{
		Sets:   s.sets,
		NextID: s.nextID,
	}

	file, err := os.Create(s.filePath)
	if err != nil {
		return fmt.Errorf("cannot create file: %s", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
func (s *fileStorage) Restore() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("cannot open file: %s", err)
	}
	defer file.Close()

	var data struct {
		Sets   map[int]*LinkSet
		NextID int
	}
	err = json.NewDecoder(file).Decode(&data)
	if err != nil && err != io.EOF {
		return fmt.Errorf("cannot decode file: %s", err)
	}
	if data.Sets != nil {
		s.sets = data.Sets
	}
	if data.NextID > 0 {
		s.nextID = data.NextID
	}
	return nil
}
func normalizeURL(raw string) string {
	if len(raw) > 0 && raw[0] != 'h' {
		return "http://" + raw
	}
	return raw
}
