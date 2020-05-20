package font

import (
	"fmt"
	"testing"

	"github.com/tdewolff/test"
)

func TestWOFFError(t *testing.T) {
	var tts = []struct {
		data string
		err  string
	}{
		{"wOFF00000000\x00\x01\x00\x0000000000000000000000i00000000000\xff\xff\xff\xfc\x00\x00\x0000000000000000000", ErrInvalidFontData.Error()},
	}
	for i, tt := range tts {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, err := ParseWOFF([]byte(tt.data))
			test.T(t, err.Error(), tt.err)
		})
	}
}