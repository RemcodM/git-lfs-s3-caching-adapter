package stats

import "os"

func TotalStats(stats []SessionStats) *SessionStats {
	totalStats := NewSessionStats()
	for _, stat := range stats {
		totalStats.Add(&stat)
	}
	return totalStats
}

func UnmarkedStats(stats []SessionStats) *SessionStats {
	unmarkedStats := NewSessionStats()
	for _, stat := range stats {
		if !stat.Marked {
			unmarkedStats.Add(&stat)
		}
	}
	return unmarkedStats
}

func MarkAll(stats []SessionStats) []error {
	var errors []error
	for _, stat := range stats {
		err := stat.Mark()
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func PurgeAll() error {
	cacheStoreDir, err := cacheStoreDirectory()
	if err != nil {
		return err
	}

	os.RemoveAll(cacheStoreDir)
	return nil
}

func ReadAllSessionStats() ([]SessionStats, []error) {
	filenames, err := allStatsFiles()
	if err != nil {
		return nil, []error{err}
	}

	var stats []SessionStats
	var errors []error

	for _, filename := range filenames {
		statsData := SessionStats{
			name: filename,
		}
		err := statsData.Read()
		if err != nil {
			errors = append(errors, err)
			continue
		}
		stats = append(stats, statsData)
	}

	return stats, errors
}
