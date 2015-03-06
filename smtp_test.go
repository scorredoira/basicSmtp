package basicSmtp

import "testing"

func TestSend(t *testing.T) {
	err := Send("test", "body", "mail.foo.com:25", "user",
		"pwd", "f@f.f", []string{"foo@foo.com"}, false)
	if err != nil {
		t.Fatal(err)
	}
}
