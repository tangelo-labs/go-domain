package pagination_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	domain "github.com/tangelo-labs/go-domainkit"
	"github.com/tangelo-labs/go-domainkit/pagination"
)

func TestPagination(t *testing.T) {
	t.Run("GIVEN a page of items in reverse ordering", func(t *testing.T) {
		now := time.Now()
		items := []item{
			{id: domain.NewID(), ts: now.Add(3 * time.Hour)},
			{id: domain.NewID(), ts: now.Add(2 * time.Hour)},
			{id: domain.NewID(), ts: now.Add(time.Hour)},
			{id: domain.NewID(), ts: now},
		}

		pg := page{
			items: items,
			total: 100,
		}

		t.Run("WHEN computing token for such page THEN token ID and timestamp matches the last element in the page", func(t *testing.T) {
			continuation := pg.next()

			require.NotEmpty(t, continuation.ID.String())
			require.NotZero(t, continuation.Timestamp)

			require.Equal(t, items[len(items)-1].id, continuation.ID)
			require.Equal(t, items[len(items)-1].ts.Unix(), continuation.Timestamp.Unix())
		})

		t.Run("WHEN rebuilding the token from string THEN rebuilt token matches original token", func(t *testing.T) {
			t1 := pg.next()
			str := t1.String()
			require.NotEmpty(t, str)

			t2 := pagination.ContinuationToken{}
			t2.FromString(str)

			require.True(t, t2.Equals(t1))
		})
	})
}

type item struct {
	id domain.ID
	ts time.Time
}

type page struct {
	items []item
	total int
}

func (p page) next() pagination.ContinuationToken {
	if len(p.items) == 0 {
		return pagination.ContinuationToken{}
	}

	last := p.items[len(p.items)-1]

	return pagination.ContinuationToken{
		ID:        last.id,
		Timestamp: last.ts,
	}
}
