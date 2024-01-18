package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Stats struct {
	name                       string `json:"-"`
	ObjectsPulled              uint64 `json:"objects_pulled"`
	ObjectsPushed              uint64 `json:"objects_pushed"`
	CacheHits                  uint64 `json:"cache_hits"`
	CacheMisses                uint64 `json:"cache_misses"`
	CacheErrors                uint64 `json:"cache_errors"`
	CacheAddedDuringPull       uint64 `json:"cache_added_during_pull"`
	CacheAddedDuringPush       uint64 `json:"cache_added_during_push"`
	BytesTransferredFromCache  uint64 `json:"bytes_transferred_from_cache"`
	BytesTransferredToCache    uint64 `json:"bytes_transferred_to_cache"`
	BytesTransferredFromRemote uint64 `json:"bytes_transferred_from_remote"`
	BytesTransferredToRemote   uint64 `json:"bytes_transferred_to_remote"`
	Sessions                   uint64 `json:"sessions"`
	Marked                     bool   `json:"marked"`
}

func newStats() *Stats {
	return &Stats{
		name:                       "",
		ObjectsPulled:              0,
		ObjectsPushed:              0,
		CacheHits:                  0,
		CacheMisses:                0,
		CacheErrors:                0,
		CacheAddedDuringPull:       0,
		CacheAddedDuringPush:       0,
		BytesTransferredFromCache:  0,
		BytesTransferredToCache:    0,
		BytesTransferredFromRemote: 0,
		BytesTransferredToRemote:   0,
		Sessions:                   0,
		Marked:                     false,
	}
}

func NewSessionStats() *Stats {
	stats := newStats()
	stats.Sessions = 1
	return stats
}

func NewCollectedStats() (*Stats, error) {
	stats := newStats()
	err := stats.generateName("collected-stats")
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *Stats) Add(other *Stats) {
	s.ObjectsPulled += other.ObjectsPulled
	s.ObjectsPushed += other.ObjectsPushed
	s.CacheHits += other.CacheHits
	s.CacheMisses += other.CacheMisses
	s.CacheErrors += other.CacheErrors
	s.CacheAddedDuringPull += other.CacheAddedDuringPull
	s.CacheAddedDuringPush += other.CacheAddedDuringPush
	s.BytesTransferredFromCache += other.BytesTransferredFromCache
	s.BytesTransferredToCache += other.BytesTransferredToCache
	s.BytesTransferredFromRemote += other.BytesTransferredFromRemote
	s.BytesTransferredToRemote += other.BytesTransferredToRemote
	s.Sessions += other.Sessions
}

func (s *Stats) IsZero() bool {
	return s.ObjectsPulled == 0 &&
		s.ObjectsPushed == 0 &&
		s.CacheHits == 0 &&
		s.CacheMisses == 0 &&
		s.CacheErrors == 0 &&
		s.CacheAddedDuringPull == 0 &&
		s.CacheAddedDuringPush == 0 &&
		s.BytesTransferredFromCache == 0 &&
		s.BytesTransferredToCache == 0 &&
		s.BytesTransferredFromRemote == 0 &&
		s.BytesTransferredToRemote == 0 &&
		s.Sessions == 0
}

func (s *Stats) Mark() error {
	s.Marked = true
	return s.Save()
}

func (s *Stats) Save() error {
	cacheStoreDir, err := requireCacheStoreDirectory()
	if err != nil {
		return err
	}

	if s.name == "" {
		err := s.generateName("stats")
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(fmt.Sprintf("%s/%s", cacheStoreDir, s.name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(s)
	if err != nil {
		return err
	}

	return nil
}

func (s *Stats) Read() error {
	cacheStoreDir, err := requireCacheStoreDirectory()
	if err != nil {
		return err
	}

	file, err := os.Open(fmt.Sprintf("%s/%s", cacheStoreDir, s.name))
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(s)
	if err != nil {
		return err
	}

	return nil
}

func (s *Stats) Delete() error {
	cacheStoreDir, err := requireCacheStoreDirectory()
	if err != nil {
		return err
	}

	if s.name == "" {
		return fmt.Errorf("cannot delete stats without a name")
	}

	return os.Remove(fmt.Sprintf("%s/%s", cacheStoreDir, s.name))
}

func (s *Stats) generateName(prefix string) error {
	suffix, err := randomHex(8)
	if err != nil {
		return err
	}
	s.name = fmt.Sprintf("%s-%d-%s.json", prefix, time.Now().UTC().Unix(), suffix)
	return nil
}
