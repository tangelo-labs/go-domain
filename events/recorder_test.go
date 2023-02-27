package events_test

import (
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	domain "github.com/tangelo-labs/go-domainkit"
	"github.com/tangelo-labs/go-domainkit/events"
)

func TestBaseRecorder(t *testing.T) {
	t.Run("GIVEN an empty recorder instance WHEN multiple goroutines write at the same time THEN every event is recorded correctly", func(t *testing.T) {
		gn := gofakeit.Number(5, 25)
		en := gofakeit.Number(5, 50)
		recorder := events.BaseRecorder{}

		var wg sync.WaitGroup
		for i := 0; i < gn; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for j := 0; j < en; j++ {
					recorder.Record(domain.NewID())
				}
			}()
		}

		wg.Wait()
		require.Len(t, recorder.Changes(), gn*en)

		t.Run("GIVEN a non-empty recorder instance WHEN cleaning THEN recorder is now empty", func(t *testing.T) {
			recorder.ClearChanges()
			require.Empty(t, recorder.Changes())
		})
	})
}
