package collector

// The purpose of this test is to check whether the given regular pattern
// captures cases from the application logs?
// They are not complete nor comprehensive.
// They also don't test the negative path.
// Fell free to make them better.

import (
	"math"
	"testing"

	l "github.com/peimanja/artifactory_exporter/logger"
)

var fakeExporter = Exporter{
	logger: l.New(
		l.Config{
			Format: l.FormatDefault,
			Level:  "debug",
		},
	),
}

const float64EqualityThreshold = 1e-6

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

func TestConvNum(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{
			input: `8 bytes`,
			want:  8,
		},
		{
			input: `8,888.88 MB`,
			want:  9320666234.879999,
		},
		{
			input: `88.88 GB`,
			want:  95434173317.119995,
		},
		{
			input: `888.88 GB`,
			want:  954427632517.119995,
		},
	}
	for _, tc := range tests {
		got, err := fakeExporter.convArtiToPromNumber(tc.input)
		if err != nil {
			t.Fatalf(`An error '%v' occurred during conversion.`, err)
		}
		if !almostEqual(tc.want, got) {
			t.Fatalf(`Want %f, but got %f.`, tc.want, got)
		}
	}
}

func TestConvFileStoreData(t *testing.T) {
	tests := []struct {
		input string
		want  []float64
	}{
		{
			input: `3.33 TB (3.3%)`,
			want:  []float64{3661373720494.080078, 0.033},
		},
		{
			input: `6.66 TB (6.66%)`,
			want:  []float64{7322747440988.160156, 0.0666},
		},
		{
			input: `11.11 TB (11.1%)`,
			want:  []float64{12215574184591.359375, 0.111},
		},
		{
			input: `99.99 TB (99.99%)`,
			want:  []float64{109940167661322.234375, 0.9999},
		},
		{
			input: `499.76 GB`,
			want:  []float64{536613213962.23999, 0},
		},
		{
			input: `4.82 GB (0.96%)`,
			want:  []float64{5175435591.68, 0.0096},
		},
		{
			input: `494.94 GB (99.04%)`,
			want:  []float64{531437778370.559998, 0.9904},
		},
	}
	for _, tc := range tests {
		gotSize, gotPercent, err := fakeExporter.convArtiToPromFileStoreData(tc.input)
		if err != nil {
			t.Fatalf(`An error '%v' occurred during conversion.`, err)
		}
		wantSize := tc.want[0]
		if !almostEqual(wantSize, gotSize) {
			t.Fatalf(`Problem with size. Want %f, but got %f.`, wantSize, gotSize)
		}
		wantPercent := tc.want[1]
		if !almostEqual(wantPercent, gotPercent) {
			t.Fatalf(`Problem with percentage. Want %f, but got %f.`, wantPercent, gotPercent)
		}
	}
}
