package model_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redjack/marionette/plugins/model"
)

func TestParseSleepDistribution(t *testing.T) {
	t.Run("http_timings", func(t *testing.T) {
		if dist, err := model.ParseSleepDistribution(
			`{'0.025' : 0.25,
             '0.050' : 0.25,
             '0.075' : 0.25,
             '0.1'   : 0.25}`); err != nil {
			t.Fatal(err)
		} else if diff := cmp.Diff(dist, map[float64]float64{
			0.025: 0.25,
			0.050: 0.25,
			0.075: 0.25,
			0.1:   0.25,
		}); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("web_sess", func(t *testing.T) {
		if dist, err := model.ParseSleepDistribution("{'5.0' : 1.0}"); err != nil {
			t.Fatal(err)
		} else if diff := cmp.Diff(dist, map[float64]float64{5.0: 1.0}); diff != "" {
			t.Fatal(diff)
		}
	})
}
