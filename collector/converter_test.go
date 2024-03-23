package collector

// The purpose of this test is to check whether the given regular pattern
// captures cases from the application logs?
// They are not complete nor comprehensive.
// They also don't test the negative path.
// Fell free to make them better.

import (
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
			want:  9319743488,
		},
		{
			input: `88.88 GB`,
			want:  94489280512,
		},
		{
			input: `888.88 GB`,
			want:  953482739712,
		},
	}
	for _, tc := range tests {
		got, err := fakeExporter.convArtiToPromNumber(tc.input)
		if err != nil {
			t.Fatalf(`An error '%v' occurred during conversion.`, err)
		}
		if tc.want != got {
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
			want:  []float64{3298534883328, 0.03},
		},
		{
			input: `6.66 TB (6.66%)`,
			want:  []float64{6597069766656, 0.06},
		},
		{
			input: `11.11 TB (11.1%)`,
			want:  []float64{12094627905536, 0.11},
		},
		{
			input: `99.99 TB (99.99%)`,
			want:  []float64{108851651149824, 0.99},
		},
		{
			input: `499.76 GB`,
			want:  []float64{535797170176, 0},
		},
		{
			input: `4.82 GB (0.96%)`,
			want:  []float64{4294967296, 0},
		},
		{
			input: `494.94 GB (99.04%)`,
			want:  []float64{530428461056, 0.99},
		},
	}
	for _, tc := range tests {
		gotSize, gotPercent, err := fakeExporter.convArtiToPromFileStoreData(tc.input)
		if err != nil {
			t.Fatalf(`An error '%v' occurred during conversion.`, err)
		}
		wantSize := tc.want[0]
		if wantSize != gotSize {
			t.Fatalf(`Problem with size. Want %f, but got %f.`, wantSize, gotSize)
		}
		wantPercent := tc.want[1]
		if wantPercent != gotPercent {
			t.Fatalf(`Problem with percentage. Want %f, but got %f.`, wantPercent, gotPercent)
		}
	}
}
