package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type SessionStats struct {
	name                       *string `json:"-"`
	ObjectsPulled              uint64  `json:"objects_pulled"`
	ObjectsPushed              uint64  `json:"objects_pushed"`
	CacheHits                  uint64  `json:"cache_hits"`
	CacheMisses                uint64  `json:"cache_misses"`
	CacheErrors                uint64  `json:"cache_errors"`
	CacheAddedDuringPull       uint64  `json:"cache_added_during_pull"`
	CacheAddedDuringPush       uint64  `json:"cache_added_during_push"`
	BytesTransferredFromCache  uint64  `json:"bytes_transferred_from_cache"`
	BytesTransferredToCache    uint64  `json:"bytes_transferred_to_cache"`
	BytesTransferredFromRemote uint64  `json:"bytes_transferred_from_remote"`
	BytesTransferredToRemote   uint64  `json:"bytes_transferred_to_remote"`
	Marked                     bool    `json:"marked"`
}

func NewSessionStats() *SessionStats {
	return &SessionStats{
		name:                       nil,
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
		Marked:                     false,
	}
}

func (s *SessionStats) Add(other *SessionStats) {
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
}

func (s *SessionStats) Mark() error {
	s.Marked = true
	return s.Save()
}

func (s *SessionStats) Save() error {
	cacheStoreDir, err := requireCacheStoreDirectory()
	if err != nil {
		return err
	}

	name := s.name
	if name == nil {
		suffix, err := randomHex(8)
		if err != nil {
			return err
		}
		generated_name := fmt.Sprintf("stats-%d-%s.json", time.Now().UTC().Unix(), suffix)
		name = &generated_name
	}

	file, err := os.OpenFile(fmt.Sprintf("%s/%s", cacheStoreDir, *name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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

func (s *SessionStats) Read() error {
	cacheStoreDir, err := requireCacheStoreDirectory()
	if err != nil {
		return err
	}

	file, err := os.Open(fmt.Sprintf("%s/%s", cacheStoreDir, *s.name))
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
