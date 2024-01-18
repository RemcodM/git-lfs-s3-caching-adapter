package stats

import "os"

func CollectedStats(stats []Stats) (*Stats, error) {
	totalStats, err := NewCollectedStats()
	if err != nil {
		return nil, err
	}
	for _, stat := range stats {
		totalStats.Add(&stat)
	}
	return totalStats, nil
}

func Compact(stats []Stats) (*Stats, []error) {
	collectedStats, err := CollectedStats(stats)
	if err != nil {
		return nil, []error{err}
	}
	if !collectedStats.IsZero() {
		err := collectedStats.Save()
		if err != nil {
			return collectedStats, []error{err}
		}
	}
	return collectedStats, Purge(stats)
}

func MarkAll(stats []Stats) []error {
	var errors []error
	for _, stat := range stats {
		err := stat.Mark()
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func FilterMarked(stats []Stats) []Stats {
	var markedStats []Stats
	for _, stat := range stats {
		if stat.Marked {
			markedStats = append(markedStats, stat)
		}
	}
	return markedStats
}

func FilterUnmarked(stats []Stats) []Stats {
	var unmarkedStats []Stats
	for _, stat := range stats {
		if !stat.Marked {
			unmarkedStats = append(unmarkedStats, stat)
		}
	}
	return unmarkedStats
}

func Purge(stats []Stats) []error {
	var errors []error
	for _, stat := range stats {
		err := stat.Delete()
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

func ReadAllStats() ([]Stats, []error) {
	filenames, err := allStatsFiles()
	if err != nil {
		return nil, []error{err}
	}

	var stats []Stats
	var errors []error

	for _, filename := range filenames {
		statsData := Stats{
			name:     filename,
			Sessions: 1, // Default to 1 session, as this is the default for older stats files
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
