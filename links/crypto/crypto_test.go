package crypto_test

import (
	"testing"

	"github.com/derinil/links/links/crypto"
	"github.com/stretchr/testify/require"
)

func TestSha256Compare(t *testing.T) {
	testCases := []struct {
		name, key, seed string
		oldHash         string
		same            bool
	}{
		{
			name:    "good hash",
			key:     "test",
			seed:    "hello",
			oldHash: "491738c463d855d98ecb21a025e06b093fa08d13b382d17b93db9cd0c40318fd",
			same:    true,
		},
		{
			name:    "bad key",
			key:     "hello",
			seed:    "hello",
			oldHash: "491738c463d855d98ecb21a025e06b093fa08d13b382d17b93db9cd0c40318fd",
		},
		{
			name:    "bad seed",
			key:     "test",
			seed:    "test",
			oldHash: "491738c463d855d98ecb21a025e06b093fa08d13b382d17b93db9cd0c40318fd",
		},
		{
			name:    "empty key",
			key:     "",
			seed:    "test",
			oldHash: "491738c463d855d98ecb21a025e06b093fa08d13b382d17b93db9cd0c40318fd",
		},
		{
			name:    "empty seed",
			key:     "hello",
			oldHash: "491738c463d855d98ecb21a025e06b093fa08d13b382d17b93db9cd0c40318fd",
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			_, err := crypto.Sha256(c.key, c.seed)
			require.Nil(t, err)
			ok, err := crypto.CompareSha256(c.key, c.seed, c.oldHash)
			require.Nil(t, err)
			require.Equal(t, c.same, ok)
		})
	}
}
