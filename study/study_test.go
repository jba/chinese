package study

import "testing"

func TestChooseWord(t *testing.T) {
	lexicon := map[string][]*Word{
		"one": []*Word{{English: "oink", Pinyin: "OINK"}},
	}
	for _, test := range []struct {
		in      string
		want    string
		wantKey string
		wantVal string
	}{
		{"foo", "foo", "", ""},
		{":zero", "???", "zero", "???"},
		{":one", "oink", "one", "OINK"},
		{":one2", "oink", "one2", "OINK"},
	} {
		bs := map[string]string{}
		got := chooseWord(test.in, lexicon, bs)
		if got != test.want {
			t.Errorf("%s: got %q, want %q", test.in, got, test.want)
		}
		if test.wantKey == "" && len(bs) != 0 {
			t.Errorf("%s: got %d bindings, wanted none", test.in, len(bs))
		}
		if test.wantKey != "" && (len(bs) != 1 || bs[test.wantKey] != test.wantVal) {
			t.Errorf("want %s => %s, got %+v", test.wantKey, test.wantVal, bs)
		}
	}
}
