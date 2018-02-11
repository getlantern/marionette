package tg

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/redjack/marionette"
)

type HTTPContentLengthCipher struct{}

func NewHTTPContentLengthCipher() *HTTPContentLengthCipher {
	return &HTTPContentLengthCipher{}
}

func (c *HTTPContentLengthCipher) Key() string {
	return "CONTENT-LENGTH"
}

func (c *HTTPContentLengthCipher) Capacity() int {
	return 0
}

func (c *HTTPContentLengthCipher) Encrypt(fsm marionette.FSM, template string, plaintext []byte) (ciphertext []byte, err error) {
	a := strings.SplitN(template, "\r\n\r\n", 2)
	if len(a) == 1 {
		return []byte("0"), nil
	}
	return []byte(strconv.Itoa(len(a[1]))), nil
}

func (c *HTTPContentLengthCipher) Decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	return nil, nil
}

func parseHTTPHeader(header_name, msg string) string {
	lines := strings.Split(msg, "\r\n")
	for _, line := range lines[1 : len(lines)-2] {
		if a := strings.SplitN(line, ": ", 2); a[0] == header_name {
			if len(a) > 1 {
				return a[1]
			}
			return ""
		}
	}
	return ""
}

func parseHTTPRequest(msg string) map[string]string {
	if !strings.HasPrefix(msg, "GET") {
		return nil
	} else if !strings.HasSuffix(msg, "\r\n\r\n") {
		return nil
	}

	lines := lineBreakRegex.Split(msg, -1)
	segments := strings.Split(lines[0][:len(lines[0])-9], "/")

	if strings.HasPrefix(msg, "GET http") {
		return map[string]string{"URL": strings.Join(segments[3:], "/")}
	}
	return map[string]string{"URL": strings.Join(segments[1:], "/")}
}

func parseHTTPResponse(msg string) map[string]string {
	if !strings.HasPrefix(msg, "HTTP") {
		return nil
	}

	m := make(map[string]string)
	m["CONTENT-LENGTH"] = parseHTTPHeader("Content-Length", msg)
	m["COOKIE"] = parseHTTPHeader("Cookie", msg)
	if a := strings.Split(msg, "\r\n\r\n"); len(a) > 1 {
		m["HTTP-RESPONSE-BODY"] = a[1]
	} else {
		m["HTTP-RESPONSE-BODY"] = ""
	}

	if m["CONTENT-LENGTH"] != strconv.Itoa(len(m["HTTP-RESPONSE-BODY"])) {
		return nil
	}
	return m
}

var lineBreakRegex = regexp.MustCompile(`\r\n`)
