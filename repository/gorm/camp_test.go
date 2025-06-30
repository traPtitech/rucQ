package gorm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestCreateCamp(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		dateStart := random.Time(t)
		dateEnd := dateStart.Add(time.Duration(random.PositiveInt(t)))
		camp := model.Camp{
			DisplayID:          random.AlphaNumericString(t, 10),
			Name:               random.AlphaNumericString(t, 20),
			Description:        random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
			DateStart:          dateStart,
			DateEnd:            dateEnd,
		}
		err := r.CreateCamp(&camp)

		assert.NoError(t, err)
	})
}

func TestGetCamps(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp1 := mustCreateCamp(t, r)
		camp2 := mustCreateCamp(t, r)
		camps, err := r.GetCamps()

		assert.NoError(t, err)
		assert.Len(t, camps, 2)
		assert.Contains(t, camps, camp1)
		assert.Contains(t, camps, camp2)
	})
}
