package ibsync

import "testing"

func TestKey(t *testing.T) {
	tests := []struct {
		input    []any
		expected string
	}{
		{[]any{"key1", "key2", "key3"}, "key1" + sep + "key2" + sep + "key3"},
		{[]any{}, ""},
		{[]any{42, "test"}, "42" + sep + "test"},
	}

	for _, test := range tests {
		result := Key(test.input...)
		if result != test.expected {
			t.Errorf("Key(%v) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

func TestEncode(t *testing.T) {
	pnl := Pnl{Account: "12345", ModelCode: "ABC", DailyPNL: 100.0, UnrealizedPnl: 50.0, RealizedPNL: 75.0}
	encoded := Encode(pnl)

	if encoded == "" {
		t.Errorf("encoded data should not be empty")
	}
}

func TestDecode(t *testing.T) {
	pnl := Pnl{Account: "12345", ModelCode: "ABC", DailyPNL: 100.0, UnrealizedPnl: 50.0, RealizedPNL: 75.0}
	encoded := Encode(pnl)

	var decoded Pnl
	err := Decode(&decoded, encoded)
	if err != nil {
		t.Errorf("expected no error during decoding, got %v", err)
	}
	if decoded != pnl {
		t.Errorf("decoded Pnl should match original: got %+v, want %+v", decoded, pnl)
	}
}

func TestJoin(t *testing.T) {
	strs := []string{"Hello", "World"}
	result := Join(strs...)
	expected := "Hello" + sep + "World"
	if result != expected {
		t.Errorf("Join did not return the expected result: got %v, want %v", result, expected)
	}
}

func TestSplit(t *testing.T) {
	str := "Hello" + sep + "World"
	result := Split(str)
	expected := []string{"Hello", "World"}
	if len(result) != len(expected) {
		t.Errorf("Split did not return expected slices: got %v, want %v", result, expected)
	}
}
