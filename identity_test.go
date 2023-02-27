package domain_test

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	domain "github.com/tangelo-labs/go-domain"
)

func TestIDNew(t *testing.T) {
	t.Run("GIVEN an ID generator WHEN generating multiple ids in parallel THEN no collisions are detected", func(t *testing.T) {
		numCPU := runtime.NumCPU()
		if numCPU < 2 {
			t.Skip("this tests requires more than one CPU")
		}

		runtime.GOMAXPROCS(numCPU)

		max := 10000
		var generated sync.Map

		var wg sync.WaitGroup
		{
			for i := 1; i <= max; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					id := domain.NewID().String()
					runtime.Gosched()

					generated.Store(id, struct{}{})
				}()
			}
		}
		wg.Wait()

		itemsCount := 0
		generated.Range(func(key, value interface{}) bool {
			itemsCount++

			return true
		})

		require.Equal(t, max, itemsCount)
	})

	t.Run("GIVEN an ID generator WHEN generating multiple ids consecutively THEN no collisions are detected", func(t *testing.T) {
		generated := make(map[string]struct{})
		max := 1000000

		for i := 0; i < max; i++ {
			generated[domain.NewID().String()] = struct{}{}
		}

		require.Equal(t, max, len(generated))
	})
}

func TestSetIDGenerator(t *testing.T) {
	t.Run("GIVEN a generator setter function", func(t *testing.T) {
		t.Run("WHEN defining an UUID algorithm THEN generated ids have uuid format", func(t *testing.T) {
			domain.SetIDGenerator(domain.UUIDGenerator)

			id := domain.NewID().String()

			require.NotEmpty(t, id)
			require.Equal(t, 36, len(id))
		})

		t.Run("WHEN defining an ULID algorithm THEN generated ids have ulid format", func(t *testing.T) {
			domain.SetIDGenerator(domain.ULIDGenerator)

			id := domain.NewID().String()

			require.NotEmpty(t, id)
			require.Equal(t, 26, len(id))
		})

		t.Run("WHEN defining a fixed generator THEN generated ids have the same value", func(t *testing.T) {
			domain.SetIDGenerator(func() domain.ID {
				return "fixed"
			})

			require.Equal(t, "fixed", domain.NewID().String())
			require.Equal(t, "fixed", domain.NewID().String())
			require.Equal(t, "fixed", domain.NewID().String())
		})
	})
}

func TestIDBasic(t *testing.T) {
	var (
		idA domain.ID = "  "
		idB domain.ID
	)

	require.NotEmpty(t, idA)
	require.False(t, idA.IsEmpty())

	require.Empty(t, idB)
	require.True(t, idB.IsEmpty())

	require.False(t, idA.Equals(idB))
	require.False(t, idB.Equals(idA))

	makeA := domain.ID(idA.String())
	makeB := domain.ID(idB.String())

	require.True(t, idA.Equals(makeA))
	require.True(t, idB.Equals(makeB))
}

func TestIDMarshalingBinary(t *testing.T) {
	t.Run("GIVEN two random ids binary encoded", func(t *testing.T) {
		idA := domain.NewID()
		idB := domain.NewID()

		a, err := idA.MarshalBinary()
		require.NoError(t, err)
		require.NotEmpty(t, a)

		b, err := idB.MarshalBinary()
		require.NoError(t, err)
		require.NotEmpty(t, b)

		t.Run("WHEN unmarshaling them back THEN they have the same original values", func(t *testing.T) {
			var aR domain.ID
			require.NoError(t, aR.UnmarshalBinary(a))
			require.True(t, idA.Equals(aR))

			var bR domain.ID
			require.NoError(t, bR.UnmarshalBinary(b))
			require.True(t, idB.Equals(bR))
		})
	})
}

func TestIDMarshalingText(t *testing.T) {
	t.Run("GIVEN two random ids text encoded", func(t *testing.T) {
		idA := domain.NewID()
		idB := domain.NewID()

		a, err := idA.MarshalText()
		require.NoError(t, err)
		require.NotEmpty(t, a)

		b, err := idB.MarshalText()
		require.NoError(t, err)
		require.NotEmpty(t, b)

		t.Run("WHEN unmarshaling them back THEN they have the same original values", func(t *testing.T) {
			var aR domain.ID
			require.NoError(t, aR.UnmarshalText(a))
			require.True(t, idA.Equals(aR))

			var bR domain.ID
			require.NoError(t, bR.UnmarshalText(b))
			require.True(t, idB.Equals(bR))
		})
	})
}

func TestIDMarshalingJSON(t *testing.T) {
	t.Run("GIVEN two random ids json encoded", func(t *testing.T) {
		idA := domain.NewID()
		idB := domain.NewID()

		a, err := idA.MarshalJSON()
		require.NoError(t, err)
		require.NotEmpty(t, a)

		b, err := idB.MarshalJSON()
		require.NoError(t, err)
		require.NotEmpty(t, b)

		t.Run("WHEN unmarshaling them back THEN they have the same original values", func(t *testing.T) {
			var aR domain.ID
			require.NoError(t, aR.UnmarshalJSON(a))
			require.True(t, idA.Equals(aR))

			var bR domain.ID
			require.NoError(t, bR.UnmarshalJSON(b))
			require.True(t, idB.Equals(bR))
		})
	})
}

func TestIDs_Contains(t *testing.T) {
	tests := []struct {
		ids      domain.IDs
		contains domain.ID
		want     bool
	}{
		{
			ids:      []domain.ID{"a", "b", "c"},
			contains: "a",
			want:     true,
		},
		{
			ids:      []domain.ID{"a", "b", "c"},
			contains: "d",
			want:     false,
		},

		{
			ids:      []domain.ID{"a", "b", "c"},
			contains: "",
			want:     false,
		},

		{
			ids:      []domain.ID{},
			contains: "a",
			want:     false,
		},

		{
			ids:      []domain.ID{},
			contains: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%+v contains %s", tt.ids, tt.contains), func(t *testing.T) {
			if got := tt.ids.Contains(tt.contains); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIDs_Subtract(t *testing.T) {
	tests := []struct {
		name  string
		left  domain.IDs
		right domain.IDs
		want  domain.IDs
	}{
		{
			name:  "subtracts two empty sets",
			left:  nil,
			right: nil,
			want:  nil,
		},
		{
			name:  "{a,a,b,b,c,c} sub {a,b,c} = {}",
			left:  []domain.ID{"a", "a", "b", "b", "c", "c"},
			right: []domain.ID{"a", "b", "c"},
			want:  []domain.ID{},
		},
		{
			name:  "{a,b,c,d,d} sub {b,d} = {a,c}",
			left:  []domain.ID{"a", "b", "c", "d", "d"},
			right: []domain.ID{"b", "d"},
			want:  []domain.ID{"a", "c"},
		},
		{
			name:  "{a,b,c,d,d} sub {} = {a,b,c,d,d}",
			left:  []domain.ID{"a", "b", "c", "d", "d"},
			right: []domain.ID{},
			want:  []domain.ID{"a", "b", "c", "d", "d"},
		},
		{
			name:  "{a,b,c,d} sub {e} = {a,b,c,d}",
			left:  []domain.ID{"a", "b", "c", "d"},
			right: []domain.ID{"e"},
			want:  []domain.ID{"a", "b", "c", "d"},
		},
		{
			name:  "{e} sub {a,b,c,d} = {e}",
			left:  []domain.ID{"e"},
			right: []domain.ID{"a", "b", "c", "d"},
			want:  []domain.ID{"e"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.left.Subtract(tt.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subtract() = %v, want %v", got, tt.want)
			}
		})
	}
}
