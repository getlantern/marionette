package marionette

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/fte"
	"go.uber.org/zap"
)

func init() {
	marionette.RegisterPlugin("tg", "send", Send)
}

func Send(fsm marionette.FSM, args []interface{}) (success bool, err error) {
	logger := marionette.Logger.With(zap.String("party", fsm.Party()))

	if len(args) < 1 {
		return false, errors.New("tg.send: not enough arguments")
	}

	grammar, ok := args[0].(string)
	if !ok {
		return false, errors.New("tg.send: invalid grammar argument type")
	}

	grammarTemplates := templates[grammar]
	if len(templates[grammar]) == 0 {
		return false, fmt.Errorf("tg.send: grammar has no templates: %q", grammar)
	}
	ciphertext := grammarTemplates[rand.Intn(len(grammarTemplates))]

	for _, handler := range tgConfigs[grammar].handlers {
		if ciphertext, err = execute_handler_sender(fsm, grammar, handler, ciphertext); err != nil {
			return false, fmt.Errorf("tg.send: execute handler sender: %q", err)
		}
	}

	logger.Debug("tg.send: writing cell data")

	// Write to outgoing connection.
	if _, err := fsm.Conn().Write([]byte(ciphertext)); err != nil {
		return false, err
	}

	logger.Debug("tg.send: cell data written")
	return true, nil
}

func do_embed(template, handler_key, value string) string {
	return strings.Replace(template, "%%"+handler_key+"%%", value, -1)
}

func do_unembed(grammar, ciphertext, handler_key string) string {
	m := parse(grammar, ciphertext)
	return m[handler_key]
}

func execute_handler_sender(fsm marionette.FSM, grammar string, handler *tgHandler, template string) (string, error) {
	// Encode data from streams if there is capacity in the handler.
	var data []byte
	if capacity := handler.cipher.capacity(); capacity > 0 {
		cell := fsm.StreamSet().Dequeue(handler.cipher.capacity())
		if cell == nil {
			cell = marionette.NewCell(0, 0, 0, marionette.NORMAL)
		}

		// Assign ids and marshal to bytes.
		cell.UUID, cell.InstanceID = fsm.UUID(), fsm.InstanceID()

		var err error
		if data, err = cell.MarshalBinary(); err != nil {
			return "", err
		}
	}

	value_to_embed, err := handler.cipher.encrypt(fsm, template, data)
	if err != nil {
		return "", err
	}
	return do_embed(template, handler.name, string(value_to_embed)), nil
}

func execute_handler_receiver(fsm marionette.FSM, grammar, handler_key, ciphertext string) (string, error) {
	handler_key_value := do_unembed(grammar, ciphertext, handler_key)
	h := tgConfigs[grammar].handler(handler_key)

	buf, err := h.cipher.decrypt(fsm, []byte(handler_key_value))
	return string(buf), err
}

type tgConfig struct {
	grammar  string
	handlers []*tgHandler
}

func (c *tgConfig) handler(name string) *tgHandler {
	for _, h := range c.handlers {
		if h.name == name {
			return h
		}
	}
	return nil
}

type tgHandler struct {
	name   string
	cipher tgCipher
}

type tgCipher interface {
	capacity() int
	encrypt(fsm marionette.FSM, template string, plaintext []byte) (ciphertext []byte, err error)
	decrypt(fsm marionette.FSM, cipher []byte) (plaintext []byte, err error)
}

var _ tgCipher = &rankerCipher{}

type rankerCipher struct {
	regex   string
	encoder *fte.DFA
}

func newRankerCipher(regex string, msg_len int) *rankerCipher {
	return &rankerCipher{
		regex:   regex,
		encoder: fte.NewDFA(regex, msg_len),
	}
}

func (c *rankerCipher) capacity() int {
	return c.encoder.Capacity()
}

func (c *rankerCipher) encrypt(fsm marionette.FSM, template string, data []byte) (ciphertext []byte, err error) {
	rank := &big.Int{}
	rank.SetBytes(data)

	ret, err := c.encoder.Unrank(rank)
	if err != nil {
		return nil, err
	}
	return []byte(ret), nil
}

func (c *rankerCipher) decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	rank, err := c.encoder.Rank(string(ciphertext))
	if err != nil {
		return nil, err
	}
	return rank.Bytes(), nil
}

var _ tgCipher = &fteCipher{}

type fteCipher struct {
	regex       string
	useCapacity bool
	encoder     *fte.Cipher
}

func newFTECipher(regex string, msg_len int, useCapacity bool) *fteCipher {
	return &fteCipher{
		regex:       regex,
		useCapacity: useCapacity,
		encoder:     fte.NewCipher(regex),
	}
}

func (c *fteCipher) capacity() int {
	if !c.useCapacity && strings.HasSuffix(c.regex, ".+") {
		return (1 << 18)
	}
	return c.encoder.Capacity() - fte.COVERTEXT_HEADER_LEN_CIPHERTTEXT - fte.CTXT_EXPANSION
}

func (c *fteCipher) encrypt(fsm marionette.FSM, template string, data []byte) (ciphertext []byte, err error) {
	return c.encoder.Encrypt(data)
}

func (c *fteCipher) decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	plaintext, _, err = c.encoder.Decrypt(ciphertext)
	return plaintext, err
}

var _ tgCipher = &httpContentLengthCipher{}

type httpContentLengthCipher struct{}

func (c *httpContentLengthCipher) capacity() int {
	return 0
}

func (c *httpContentLengthCipher) encrypt(fsm marionette.FSM, template string, plaintext []byte) (ciphertext []byte, err error) {
	a := strings.SplitN(template, "\r\n\r\n", 2)
	if len(a) == 1 {
		return []byte("0"), nil
	}
	return []byte(strconv.Itoa(len(a[1]))), nil
}

func (c *httpContentLengthCipher) decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	return nil, nil
}

var _ tgCipher = &pop3ContentLengthCipher{}

type pop3ContentLengthCipher struct{}

func (c *pop3ContentLengthCipher) capacity() int {
	return 0
}

func (c *pop3ContentLengthCipher) encrypt(fsm marionette.FSM, template string, plaintext []byte) (ciphertext []byte, err error) {
	a := strings.SplitN(template, "\n", 2)
	if len(a) == 1 {
		return []byte("0"), nil
	}
	return []byte(strconv.Itoa(len(a[1]))), nil
}

func (c *pop3ContentLengthCipher) decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	return nil, nil
}

var _ tgCipher = &setFTPPasvXCipher{}

type setFTPPasvXCipher struct{}

func (c *setFTPPasvXCipher) capacity() int {
	return 0
}

func (c *setFTPPasvXCipher) encrypt(fsm marionette.FSM, template string, plaintext []byte) (ciphertext []byte, err error) {
	i := fsm.Var("ftp_pasv_port").(int)
	return []byte(strconv.Itoa(i / 256)), nil
}

func (c *setFTPPasvXCipher) decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	i, _ := strconv.Atoi(string(ciphertext))
	fsm.SetVar("ftp_pasv_port_x", i)
	return nil, nil
}

var _ tgCipher = &setFTPPasvYCipher{}

type setFTPPasvYCipher struct{}

func (c *setFTPPasvYCipher) capacity() int {
	return 0
}

func (c *setFTPPasvYCipher) encrypt(fsm marionette.FSM, template string, plaintext []byte) (ciphertext []byte, err error) {
	i := fsm.Var("ftp_pasv_port").(int)
	return []byte(strconv.Itoa(i % 256)), nil
}

func (c *setFTPPasvYCipher) decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	ftp_pasv_port_x := fsm.Var("ftp_pasv_port").(int)
	ftp_pasv_port_y, _ := strconv.Atoi(string(ciphertext))

	fsm.SetVar("ftp_pasv_port", ftp_pasv_port_x*256+ftp_pasv_port_y)
	return nil, nil
}

var _ tgCipher = &setDNSTransactionIDCipher{}

type setDNSTransactionIDCipher struct{}

func (c *setDNSTransactionIDCipher) capacity() int {
	return 0
}

func (c *setDNSTransactionIDCipher) encrypt(fsm marionette.FSM, template string, plaintext []byte) (ciphertext []byte, err error) {
	var dns_transaction_id string
	if v := fsm.Var("dns_transaction_id"); v != nil {
		dns_transaction_id = v.(string)
	} else {
		dns_transaction_id = string([]rune{rune(rand.Intn(253) + 1), rune(rand.Intn(253) + 1)})
		fsm.SetVar("dns_transaction_id", dns_transaction_id)
	}
	return []byte(dns_transaction_id), nil
}

func (c *setDNSTransactionIDCipher) decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	fsm.SetVar("dns_transaction_id", string(ciphertext))
	return nil, nil
}

var _ tgCipher = &setDNSDomainCipher{}

type setDNSDomainCipher struct{}

func (c *setDNSDomainCipher) capacity() int {
	return 0
}

func (c *setDNSDomainCipher) encrypt(fsm marionette.FSM, template string, plaintext []byte) (ciphertext []byte, err error) {
	var dns_domain string
	if v := fsm.Var("dns_domain"); v != nil {
		dns_domain = v.(string)
	} else {
		available := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
		tlds := []string{"com", "net", "org"}

		buf := make([]rune, rand.Intn(60)+3+1)
		buf[0] = rune(len(buf) - 1) // name length
		for i := 1; i < len(buf); i++ {
			buf[i] = available[rand.Intn(len(available))]
		}
		buf = append(buf, 3) // tld length
		buf = append(buf, []rune(tlds[rand.Intn(len(tlds))])...)

		dns_domain = string(buf)
		fsm.SetVar("dns_domain", dns_domain)
	}
	return []byte(dns_domain), nil
}

func (c *setDNSDomainCipher) decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	fsm.SetVar("dns_domain", string(ciphertext))
	return nil, nil
}

var _ tgCipher = &setDNSIPCipher{}

type setDNSIPCipher struct{}

func (c *setDNSIPCipher) capacity() int {
	return 0
}

func (c *setDNSIPCipher) encrypt(fsm marionette.FSM, template string, plaintext []byte) (ciphertext []byte, err error) {
	var dns_ip string
	if v := fsm.Var("dns_ip"); v != nil {
		dns_ip = v.(string)
	} else {
		dns_ip = string([]rune{rune(rand.Intn(253) + 1), rune(rand.Intn(253) + 1), rune(rand.Intn(253) + 1), rune(rand.Intn(253) + 1)})
		fsm.SetVar("dns_ip", dns_ip)
	}
	return []byte(dns_ip), nil
}

func (c *setDNSIPCipher) decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	fsm.SetVar("dns_ip", string(ciphertext))
	return nil, nil
}

var amazon_msg_lens []int

func init() {
	for _, item := range []struct {
		n      int
		weight int
	}{
		{n: 2049, weight: 1},
		{n: 2052, weight: 2},
		{n: 2054, weight: 2},
		{n: 2057, weight: 3},
		{n: 2058, weight: 2},
		{n: 2059, weight: 1},
		{n: 2065, weight: 1},
		{n: 17429, weight: 1},
		{n: 3098, weight: 1},
		{n: 687, weight: 3},
		{n: 2084, weight: 1},
		{n: 42, weight: 58},
		{n: 43, weight: 107},
		{n: 9260, weight: 1},
		{n: 11309, weight: 1},
		{n: 11829, weight: 1},
		{n: 9271, weight: 1},
		{n: 6154, weight: 1},
		{n: 64, weight: 15},
		{n: 1094, weight: 1},
		{n: 12376, weight: 1},
		{n: 89, weight: 1},
		{n: 10848, weight: 1},
		{n: 5223, weight: 1},
		{n: 69231, weight: 1},
		{n: 7795, weight: 1},
		{n: 2678, weight: 1},
		{n: 8830, weight: 1},
		{n: 29826, weight: 1},
		{n: 16006, weight: 10},
		{n: 8938, weight: 1},
		{n: 17055, weight: 2},
		{n: 87712, weight: 1},
		{n: 23202, weight: 1},
		{n: 7441, weight: 1},
		{n: 17681, weight: 1},
		{n: 12456, weight: 1},
		{n: 41132, weight: 1},
		{n: 25263, weight: 6},
		{n: 689, weight: 1},
		{n: 9916, weight: 1},
		{n: 10101, weight: 2},
		{n: 1730, weight: 1},
		{n: 10948, weight: 1},
		{n: 26826, weight: 1},
		{n: 6357, weight: 1},
		{n: 13021, weight: 2},
		{n: 1246, weight: 4},
		{n: 19683, weight: 1},
		{n: 1765, weight: 1},
		{n: 1767, weight: 1},
		{n: 1768, weight: 1},
		{n: 1769, weight: 4},
		{n: 1770, weight: 6},
		{n: 1771, weight: 3},
		{n: 1772, weight: 2},
		{n: 1773, weight: 4},
		{n: 1774, weight: 4},
		{n: 1775, weight: 1},
		{n: 1776, weight: 1},
		{n: 1779, weight: 1},
		{n: 40696, weight: 1},
		{n: 767, weight: 1},
		{n: 17665, weight: 1},
		{n: 27909, weight: 1},
		{n: 12550, weight: 1},
		{n: 5385, weight: 1},
		{n: 16651, weight: 1},
		{n: 5392, weight: 1},
		{n: 26385, weight: 1},
		{n: 12056, weight: 1},
		{n: 41245, weight: 2},
		{n: 13097, weight: 1},
		{n: 15152, weight: 1},
		{n: 310, weight: 1},
		{n: 40759, weight: 1},
		{n: 9528, weight: 1},
		{n: 8000, weight: 7},
		{n: 471, weight: 1},
		{n: 15180, weight: 1},
		{n: 14158, weight: 3},
		{n: 37719, weight: 2},
		{n: 1895, weight: 1},
		{n: 31082, weight: 1},
		{n: 19824, weight: 1},
		{n: 30956, weight: 1},
		{n: 18807, weight: 1},
		{n: 11095, weight: 1},
		{n: 37756, weight: 2},
		{n: 746, weight: 1},
		{n: 10475, weight: 1},
		{n: 4332, weight: 1},
		{n: 35730, weight: 1},
		{n: 11667, weight: 1},
		{n: 16788, weight: 1},
		{n: 12182, weight: 4},
		{n: 39663, weight: 1},
		{n: 9126, weight: 1},
		{n: 35760, weight: 1},
		{n: 12735, weight: 1},
		{n: 6594, weight: 1},
		{n: 451, weight: 15},
		{n: 19402, weight: 1},
		{n: 463, weight: 3},
		{n: 10193, weight: 1},
		{n: 16853, weight: 6},
		{n: 982, weight: 1},
		{n: 15865, weight: 1},
		{n: 2008, weight: 2},
		{n: 476, weight: 1},
		{n: 13655, weight: 1},
		{n: 10213, weight: 1},
		{n: 10737, weight: 1},
		{n: 15858, weight: 1},
		{n: 2035, weight: 6},
		{n: 2039, weight: 1},
		{n: 2041, weight: 2},
	} {
		for i := 0; i < item.weight; i++ {
			amazon_msg_lens = append(amazon_msg_lens, item.n)
		}
	}
}

const MIN_PTXT_LEN = fte.COVERTEXT_HEADER_LEN_CIPHERTTEXT + fte.CTXT_EXPANSION + 32

type amazonMsgLensHandler struct {
	min_len    int
	max_len    int
	target_len int
	regex      string
	encoder    *fte.Cipher
	dfas       map[dfaKey]*fte.DFA
}

func newAmazonMsgLensHandler(regex string) *amazonMsgLensHandler {
	return &amazonMsgLensHandler{
		min_len:    MIN_PTXT_LEN,
		max_len:    1 << 18,
		target_len: 0,
		regex:      regex,
		encoder:    fte.NewCipher(regex),
		dfas:       make(map[dfaKey]*fte.DFA),
	}
}

func (h *amazonMsgLensHandler) capacity() int {
	h.target_len = amazon_msg_lens[rand.Intn(len(amazon_msg_lens))]
	if h.target_len < h.min_len {
		return 0
	} else if h.target_len > h.max_len {
		// We do this to prevent unranking really large slices
		// in practice this is probably bad since it unnaturally caps
		// our message sizes to whatever FTE can support
		h.target_len = h.max_len
		return h.max_len
	}
	n := h.target_len - fte.COVERTEXT_HEADER_LEN_CIPHERTTEXT
	n -= fte.CTXT_EXPANSION
	// n = int(ptxt_len * 8.0)-1
	return n
}

func (h *amazonMsgLensHandler) encrypt(fsm marionette.FSM, template string, plaintext []byte) (ciphertext []byte, err error) {
	if h.target_len < h.min_len || h.target_len > h.max_len {
		key := dfaKey{h.regex, h.target_len}
		dfa := h.dfas[key]
		if dfa == nil {
			dfa = fte.NewDFA(h.regex, h.target_len)
			h.dfas[key] = dfa
		}

		numWords, err := dfa.NumWordsInSlice(h.target_len)
		if err != nil {
			return nil, err
		}

		ret, err := dfa.Unrank(big.NewInt(rand.Int63n(int64(numWords))))
		if err != nil {
			return nil, err
		}
		return []byte(ret), nil
	}

	ciphertext, err = h.encoder.Encrypt(plaintext)
	if err != nil {
		return nil, err
	} else if len(ciphertext) != h.target_len {
		return nil, fmt.Errorf("Could not find ciphertext of len %d (%d)", h.target_len, len(ciphertext))
	}
	return ciphertext, nil
}

func (h *amazonMsgLensHandler) decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	if len(ciphertext) < h.min_len {
		return nil, nil
	}
	plaintext, _, err = h.encoder.Decrypt(ciphertext)
	return plaintext, err
}

type dfaKey struct {
	regex string
	n     int
}

var tgConfigs = map[string]*tgConfig{
	"http_request_keep_alive": &tgConfig{
		grammar: "http_request_keep_alive",
		handlers: []*tgHandler{
			{name: "URL", cipher: newRankerCipher(`[a-zA-Z0-9\?\-\.\&]+`, 2048)},
		},
	},

	"http_response_keep_alive": &tgConfig{
		grammar: "http_response_keep_alive",
		handlers: []*tgHandler{
			{name: "HTTP-RESPONSE-BODY", cipher: newFTECipher(".+", 128, false)},
			{name: "CONTENT-LENGTH", cipher: &httpContentLengthCipher{}},
		},
	},

	"http_request_close": &tgConfig{
		grammar:  "http_request_close",
		handlers: []*tgHandler{{name: "URL", cipher: newRankerCipher(`[a-zA-Z0-9\?\-\.\&]+`, 2048)}},
	},

	"http_response_close": &tgConfig{
		grammar: "http_response_close",
		handlers: []*tgHandler{
			{name: "HTTP-RESPONSE-BODY", cipher: newFTECipher(`.+`, 128, false)},
			{name: "CONTENT-LENGTH", cipher: &httpContentLengthCipher{}},
		},
	},

	"pop3_message_response": &tgConfig{
		grammar: "pop3_message_response",
		handlers: []*tgHandler{
			{name: "POP3-RESPONSE-BODY", cipher: newRankerCipher(`[a-zA-Z0-9]+`, 2048)},
			{name: "CONTENT-LENGTH", cipher: &pop3ContentLengthCipher{}},
		},
	},

	"pop3_password": &tgConfig{
		grammar:  "pop3_password",
		handlers: []*tgHandler{{name: "PASSWORD", cipher: newRankerCipher(`[a-zA-Z0-9]+`, 256)}},
	},

	"http_request_keep_alive_with_msg_lens": &tgConfig{
		grammar:  "http_request_keep_alive",
		handlers: []*tgHandler{{name: "URL", cipher: newFTECipher(`[a-zA-Z0-9\?\-\.\&]+`, 2048, true)}},
	},

	"http_response_keep_alive_with_msg_lens": &tgConfig{
		grammar: "http_response_keep_alive",
		handlers: []*tgHandler{
			{name: "HTTP-RESPONSE-BODY", cipher: newFTECipher(`.+`, 2048, true)},
			{name: "CONTENT-LENGTH", cipher: &httpContentLengthCipher{}},
		},
	},

	"http_amazon_request": &tgConfig{
		grammar:  "http_request_keep_alive",
		handlers: []*tgHandler{{name: "URL", cipher: newRankerCipher(`[a-zA-Z0-9\?\-\.\&]+`, 2048)}},
	},

	"http_amazon_response": &tgConfig{
		grammar: "http_response_keep_alive",
		handlers: []*tgHandler{
			{name: "HTTP-RESPONSE-BODY", cipher: newAmazonMsgLensHandler(`.+`)},
			{name: "CONTENT-LENGTH", cipher: &httpContentLengthCipher{}},
		},
	},

	"ftp_entering_passive": &tgConfig{
		grammar: "ftp_entering_passive",
		handlers: []*tgHandler{
			{name: "FTP_PASV_PORT_X", cipher: &setFTPPasvXCipher{}},
			{name: "FTP_PASV_PORT_Y", cipher: &setFTPPasvYCipher{}},
		},
	},

	"dns_request": &tgConfig{
		grammar: "dns_request",
		handlers: []*tgHandler{
			{name: "DNS_TRANSACTION_ID", cipher: &setDNSTransactionIDCipher{}},
			{name: "DNS_DOMAIN", cipher: &setDNSDomainCipher{}},
		},
	},

	"dns_response": &tgConfig{
		grammar: "dns_response",
		handlers: []*tgHandler{
			{name: "DNS_TRANSACTION_ID", cipher: &setDNSTransactionIDCipher{}},
			{name: "DNS_DOMAIN", cipher: &setDNSDomainCipher{}},
			{name: "DNS_IP", cipher: &setDNSIPCipher{}},
		},
	},
}

func parse(grammar, msg string) map[string]string {
	if strings.HasPrefix(grammar, "http_response") || grammar == "http_amazon_response" {
		return http_response_parser(msg)
	} else if strings.HasPrefix(grammar, "http_request") || grammar == "http_amazon_request" {
		return http_request_parser(msg)
	} else if strings.HasPrefix(grammar, "pop3_message_response") {
		return pop3_parser(msg)
	} else if strings.HasPrefix(grammar, "pop3_password") {
		return pop3_password_parser(msg)
	} else if strings.HasPrefix(grammar, "ftp_entering_passive") {
		return ftp_entering_passive_parser(msg)
	} else if strings.HasPrefix(grammar, "dns_request") {
		return dns_request_parser(msg)
	} else if strings.HasPrefix(grammar, "dns_response") {
		return dns_response_parser(msg)
	}
	return nil
}

var templates = map[string][]string{
	"http_request_keep_alive": []string{
		"GET http://%%SERVER_LISTEN_IP%%:8080/%%URL%% HTTP/1.1\r\nUser-Agent: marionette 0.1\r\nConnection: keep-alive\r\n\r\n",
	},
	"http_response_keep_alive": []string{
		"HTTP/1.1 200 OK\r\nContent-Length: %%CONTENT-LENGTH%%\r\nConnection: keep-alive\r\n\r\n%%HTTP-RESPONSE-BODY%%",
		"HTTP/1.1 404 Not Found\r\nContent-Length: %%CONTENT-LENGTH%%\r\nConnection: keep-alive\r\n\r\n%%HTTP-RESPONSE-BODY%%",
	},

	"http_request_keep_alive_with_msg_lens": []string{
		"GET http://%%SERVER_LISTEN_IP%%:8080/%%URL%% HTTP/1.1\r\nUser-Agent: marionette 0.1\r\nConnection: keep-alive\r\n\r\n",
	},
	"http_response_keep_alive_with_msg_lens": []string{
		"HTTP/1.1 200 OK\r\nContent-Length: %%CONTENT-LENGTH%%\r\nConnection: keep-alive\r\n\r\n%%HTTP-RESPONSE-BODY%%",
		"HTTP/1.1 404 Not Found\r\nContent-Length: %%CONTENT-LENGTH%%\r\nConnection: keep-alive\r\n\r\n%%HTTP-RESPONSE-BODY%%",
	},

	"http_amazon_request": []string{
		"GET http://%%SERVER_LISTEN_IP%%:8080/%%URL%% HTTP/1.1\r\nUser-Agent: marionette 0.1\r\nConnection: keep-alive\r\n\r\n",
	},
	"http_amazon_response": []string{
		"HTTP/1.1 200 OK\r\nContent-Length: %%CONTENT-LENGTH%%\r\nConnection: keep-alive\r\n\r\n%%HTTP-RESPONSE-BODY%%",
		"HTTP/1.1 404 Not Found\r\nContent-Length: %%CONTENT-LENGTH%%\r\nConnection: keep-alive\r\n\r\n%%HTTP-RESPONSE-BODY%%",
	},

	"http_request_close": []string{
		"GET http://%%SERVER_LISTEN_IP%%:8080/%%URL%% HTTP/1.1\r\nUser-Agent: marionette 0.1\r\nConnection: close\r\n\r\n",
	},
	"http_response_close": []string{
		"HTTP/1.1 200 OK\r\nContent-Length: %%CONTENT-LENGTH%%\r\nConnection: close\r\n\r\n%%HTTP-RESPONSE-BODY%%",
		"HTTP/1.1 404 Not Found\r\nContent-Length: %%CONTENT-LENGTH%%\r\nConnection: close\r\n\r\n%%HTTP-RESPONSE-BODY%%",
	},

	"pop3_message_response": []string{
		"+OK %%CONTENT-LENGTH%% octets\nReturn-Path: sender@example.com\nReceived: from client.example.com ([192.0.2.1])\nFrom: sender@example.com\nSubject: Test message\nTo: recipient@example.com\n\n%%POP3-RESPONSE-BODY%%\n.\n",
	},
	"pop3_password": []string{"PASS %%PASSWORD%%\n"},

	"ftp_entering_passive": []string{
		"227 Entering Passive Mode (127,0,0,1,%%FTP_PASV_PORT_X%%,%%FTP_PASV_PORT_Y%%).\n",
	},

	"dns_request": []string{
		"%%DNS_TRANSACTION_ID%%\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00%%DNS_DOMAIN%%\x00\x00\x01\x00\x01",
	},

	"dns_response": []string{
		"%%DNS_TRANSACTION_ID%%\x81\x80\x00\x01\x00\x01\x00\x00\x00\x00%%DNS_DOMAIN%%\x00\x01\x00\x01\xc0\x0c\x00\x01\x00\x01\x00\x00\x00\x02\x00\x04%%DNS_IP%%",
	},
}

func get_http_header(header_name, msg string) string {
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

func http_request_parser(msg string) map[string]string {
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

var lineBreakRegex = regexp.MustCompile(`\r\n`)

func http_response_parser(msg string) map[string]string {
	if !strings.HasPrefix(msg, "HTTP") {
		return nil
	}

	m := make(map[string]string)
	m["CONTENT-LENGTH"] = get_http_header("Content-Length", msg)
	m["COOKIE"] = get_http_header("Cookie", msg)
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

func pop3_parser(msg string) map[string]string {
	a := strings.Split(msg, "\n\n")
	if len(a) < 2 {
		return nil
	}

	body := a[1]
	if !strings.HasSuffix(body, "\n.\n") {
		return nil
	}
	body = strings.TrimSuffix(body, "\n.\n")

	return map[string]string{
		"POP3-RESPONSE-BODY": body,
		"CONTENT-LENGTH":     strconv.Itoa(len(body)),
	}
}

func pop3_password_parser(msg string) map[string]string {
	if !strings.HasSuffix(msg, "\n") {
		return nil
	}
	return map[string]string{"PASSWORD": msg[5 : len(msg)-1]}
}

func ftp_entering_passive_parser(msg string) map[string]string {
	if !strings.HasPrefix(msg, "227 Entering Passive Mode (") || !strings.HasSuffix(msg, ").\n") {
		return nil
	}

	a := strings.Split(msg, ",")
	if len(a) < 6 {
		return nil
	}

	return map[string]string{
		"FTP_PASV_PORT_X": a[4],
		"FTP_PASV_PORT_Y": a[5][:len(a[5])-3],
	}
}

func validate_dns_domain(msg string, dns_response bool) string {
	delim, splitN := "\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00", 2
	if dns_response {
		delim, splitN = "\x81\x80\x00\x01\x00\x01\x00\x00\x00\x00", 3
	}

	tmp_domain_split1 := strings.Split(msg, delim)
	if len(tmp_domain_split1) != 2 {
		return ""
	}

	tmp_domain_split2 := strings.Split(tmp_domain_split1[1], "\x00\x01\x00\x01")
	if len(tmp_domain_split2) != splitN {
		return ""
	}

	tmp_domain := tmp_domain_split2[0]

	// Check for valid prepended length
	// Remove trailing tld prepended length (1), tld (3) and trailing null (1) = 5
	if int(tmp_domain[0]) != len(tmp_domain[1:len(tmp_domain)-5]) {
		return ""
	} else if int(tmp_domain[len(tmp_domain)-5]) != 3 {
		return ""
	}

	// Check for valid TLD
	if !strings.HasSuffix(tmp_domain, "com\x00") && !strings.HasSuffix(tmp_domain, "net\x00") && !strings.HasSuffix(tmp_domain, "org\x00") {
		return ""
	}

	// Check for valid domain characters
	if !domainRegex.MatchString(tmp_domain[1 : len(tmp_domain)-5]) {
		return ""
	}

	return tmp_domain
}

var domainRegex = regexp.MustCompile(`^[\w\d]+$`)

func validate_dns_ip(msg string) string {
	tmp_ip_split := strings.Split(msg, "\x00\x01\x00\x01\xc0\x0c\x00\x01\x00\x01\x00\x00\x00\x02\x00\x04")
	if len(tmp_ip_split) != 2 {
		return ""
	}

	tmp_ip := tmp_ip_split[1]
	if len(tmp_ip) != 4 {
		return ""
	}
	return tmp_ip
}

func dns_request_parser(msg string) map[string]string {
	if !strings.Contains(msg, "\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00") {
		return nil
	}

	tmp_domain := validate_dns_domain(msg, false)
	if tmp_domain == "" {
		return nil
	}

	return map[string]string{
		"DNS_TRANSACTION_ID": msg[:2],
		"DNS_DOMAIN":         tmp_domain,
	}
}

func dns_response_parser(msg string) map[string]string {
	if !strings.Contains(msg, "\x81\x80\x00\x01\x00\x01\x00\x00\x00\x00") {
		return nil
	}

	tmp_domain := validate_dns_domain(msg, true)
	if tmp_domain == "" {
		return nil
	}

	tmp_ip := validate_dns_ip(msg)
	if tmp_ip == "" {
		return nil
	}

	return map[string]string{
		"DNS_TRANSACTION_ID": msg[:2],
		"DNS_DOMAIN":         tmp_domain,
		"DNS_IP":             tmp_ip,
	}
}
