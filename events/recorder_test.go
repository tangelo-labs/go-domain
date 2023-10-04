package events_test

import (
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domain"
	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/eventtest"
)

func TestBaseRecorder(t *testing.T) {
	t.Run("GIVEN an empty recorder instance WHEN multiple goroutines write at the same time THEN every event is recorded correctly", func(t *testing.T) {
		gn := gofakeit.Number(5, 25)
		en := gofakeit.Number(5, 50)
		recorder := &events.BaseRecorder{}

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

		t.Run("GIVEN a non-empty recorder instance", func(t *testing.T) {
			t.Run("WHEN iterating over changes THEN at least one event is non-empty", func(t *testing.T) {
				eventtest.Require.Condition(t, recorder, func(event interface{}) bool {
					if id, ok := event.(domain.ID); ok {
						return !id.IsEmpty()
					}

					return false
				})
			})

			t.Run("WHEN cleaning THEN recorder is now empty", func(t *testing.T) {
				recorder.ClearChanges()
				require.Empty(t, recorder.Changes())
			})
		})
	})
}
