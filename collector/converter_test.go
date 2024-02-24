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
			Level:  l.LevelDefault,
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
			want:  9319743488.00,
		},
		{
			input: `88.88 GB`,
			want:  94489280512.00,
		},
		{
			input: `888.88 GB`,
			want:  953482739712.00,
		},
	}
	for _, tc := range tests {
		got, err := fakeExporter.convNumArtiToProm(tc.input)
		if err != nil {
			t.Fatalf(`An error '%v' occurred during conversion.`, err)
		}
		if tc.want != got {
			t.Fatalf(`Want %f, but got %f.`, tc.want, got)
		}
	}
}

func TestConvTwoNum(t *testing.T) {
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
	}
	for _, tc := range tests {
		gotSize, gotPercent, err := fakeExporter.convTwoNumsArtiToProm(tc.input)
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
