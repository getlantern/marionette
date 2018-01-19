package marionette

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/redjack/marionette/fte"
)

func TGSendPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	logger := fsm.logger()

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

	for _, handler := range tgConfigs[grammar].Handler {
		ciphertext = execute_handler_sender(fsm, grammar, handler, ciphertext)
	}

	logger.Debug("tg.send: writing cell data")

	// Write to outgoing connection.
	if _, err := fsm.conn.Write([]byte(ciphertext)); err != nil {
		return false, err
	}

	logger.Debug("tg.send: cell data written")
	return true, nil
}

/*
def recv(channel, marionette_state, input_args):
    retval = False
    grammar = input_args[0]

    try:
        ctxt = channel.recv()

        if parser(grammar, ctxt):
            cell_str = ''
            for handler_key in conf[grammar]["handler_order"]:
                tmp_str = execute_handler_receiver(marionette_state,
                                                   grammar, handler_key, ctxt)
                if tmp_str:
                    cell_str += tmp_str

            if not cell_str:
                retval = True
            else:
                ##
                cell_obj = marionette_tg.record_layer.unserialize(cell_str)
                assert cell_obj.get_model_uuid() == marionette_state.get_local(
                    "model_uuid")

                marionette_state.set_local(
                    "model_instance_id", cell_obj.get_model_instance_id())
                ##

                if marionette_state.get_local("model_instance_id"):
                    marionette_state.get_global(
                        "multiplexer_incoming").push(cell_str)
                    retval = True
    except socket.timeout as e:
        pass
    except socket.error as e:
        pass
    except marionette_tg.record_layer.UnserializeException as e:
        pass

    if not retval:
        channel.rollback()

    return retval

*/

/*
def get_grammar_capacity(grammar):
    retval = 0
    for handler_key in conf[grammar]["handler_order"]:
        retval += conf[grammar]['handlers'][handler_key].capacity()
    retval /= 8.0
    return retval

# handler + (un)embed functions


*/

/*
def do_embed(grammar, template, handler_key, value):
    if template.count("%%" + handler_key + "%%") == 0:
        # handler not in template, no need to execute
        pass
    elif template.count("%%" + handler_key + "%%") == 1:
        template = template.replace("%%" + handler_key + "%%", value)
    else:
        # don't know how to handle >1 handlers, yet
        assert False

    return template


*/

/*
def do_unembed(grammar, ctxt, handler_key):
    parse_tree = parser(grammar, ctxt)
    return parse_tree[handler_key]


*/

func execute_handler_sender(fsm *FSM, grammar string, handler *tgHandler, template string) string {
	// Encode data from streams if there is capacity in the handler.
	var data []byte
	if capacity := handler.Capacity(); capacity > 0 {
		cell = fsm.streams.Dequeue(handler.Capacity())
		if cell == nil {
			cell = NewCell(0, 0, 0, NORMAL)
		}

		// Assign ids and marshal to bytes.
		cell.UUID, cell.InstanceID = fsm.UUID(), fsm.InstanceID
		data, _ = cell.MarshalBinary()
	}

	value_to_embed := handler.encode(fsm, template, to_embed)
	return do_embed(grammar, template, handler, value_to_embed)
}

/*
def execute_handler_receiver(marionette_state, grammar, handler_key,
                             ctxt):
    ptxt = ''

    to_execute = conf[grammar]["handlers"][handler_key]

    handler_key_value = do_unembed(grammar, ctxt, handler_key)
    ptxt = to_execute.decode(marionette_state, handler_key_value)

    return ptxt

# handlers

regex_cache_ = {}
fte_cache_ = {}
*/

type RankerCrypter struct {
	regex   string
	dfa     *fte.DFA // NOT NEEDED
	encoder *fte.DFAEncoder
}

func NewRankerCrypter(regex string, msg_len int) *RankerCrypter {
	c := &RankerCrypter{regex: regex}

	// TODO: Generate or retrieve dfa/encoder from cache.
	// regex_key = regex + str(msg_len)
	// if not regex_cache_.get(regex_key):
	//     dfa = regex2dfa.regex2dfa(regex)
	//     cDFA = fte.cDFA.DFA(dfa, msg_len)
	//     encoder = fte.dfa.DFA(cDFA, msg_len)
	//     regex_cache_[regex_key] = (dfa, encoder)
	// (self.dfa_, self.encoder_) = regex_cache_[regex_key]

	return c
}

func (c *RankerCrypter) Capacity() int {
	return c.encoder.Capacity()
}

func (c *RankerCrypter) encode(fsm *FSM, template string, data []byte) (ciphertext string) {
	to_embed_as_int = fte.bit_ops.bytes_to_long(dat)
	return self.encoder_.unrank(to_embed_as_int)
}

func (c *RankerCrypter) decode(self, marionette_state, ciphertext string) (plaintext string) {
	plaintext = self.encoder_.rank(ciphertext)
	plaintext = fte.bit_ops.long_to_bytes(plaintext, c.Capacity()/8)
	return plaintext
}

type FteCrypter struct {
	regex       string
	useCapacity bool
	encoder     *fte.Cipher
}

func NewFteCrypter(regex string, msg_len int, useCapacity bool) *FteCrypter {
	c := &FteCrypter{regex: regex}

	// TODO: Generate or retrieve dfa/encoder from cache.
	// fte_key = regex + str(msg_len)
	// if not fte_cache_.get(fte_key):
	//     dfa = regex2dfa.regex2dfa(regex)
	//     encrypter = fte.encoder.DfaEncoder(dfa, msg_len)
	//     fte_cache_[fte_key] = (dfa, encrypter)
	// (self.dfa_, self.fte_encrypter_) = fte_cache_[fte_key]

	return c
}

func (c *FteCrypter) Capacity() int {
	if !c.useCapacity && strings.HasSuffix(c.regex_, ".+") {
		return (1 << 18)
	}

	return c.fte_encrypter_.getCapacity() - fte._COVERTEXT_HEADER_LEN_CIPHERTTEXT - fte._CTXT_EXPANSION

}

func (c *FteCrypter) Encode(fsm *FSM, template string, data []byte) (ciphertext string) {
	return self.fte_encrypter_.encode(data)
}

func (c *FteCrypter) decode(self, marionette_state, ciphertext string) string {
	return self.fte_encrypter_.decode(ciphertext)
}

type HttpContentLengthCrypter struct{}

func (c *HttpContentLengthCrypter) Capacity() int {
	return 0
}

func (c *HttpContentLengthCrypter) encode(fsm *FSM, template string, data []byte) string {
	a := strings.SplitN(template, "\r\n\r\n", 2)
	if len(a) == 1 {
		return 0
	}
	return strconv.Itoa(len(a[1]))
}

func (c *HttpContentLengthCrypter) decode(fsm *FSM, ciphertext string) string {
	return ""
}

type Pop3ContentLengthCrypter struct{}

func (c *Pop3ContentLengthCrypter) Capacity() int {
	return 0
}

func (c *Pop3ContentLengthCrypter) encode(fsm *FSM, template string, data []byte) string {
	a := strings.SplitN(template, "\n", 2)
	if len(a) == 1 {
		return 0
	}
	return strconv.Itoa(len(a[1]))
}

func (c *Pop3ContentLengthCrypter) decode(fsm *FSM, ciphertext string) string {
	return ""
}

type SetFTPPasvX struct{}

func (c *SetFTPPasvX) Capacity(self) int {
	return 0
}

func (c *SetFTPPasvX) encode(fsm *FSM, template string, data []byte) string {
	i := fsm.Var("ftp_pasv_port").(int)
	return strconv.Itoa(i / 256)
}

func (c *SetFTPPasvX) decode(fsm *FSM, ciphertext string) string {
	i, _ := strconv.Atoi(ciphertext)
	fsm.SetVar("ftp_pasv_port_x", i)
	return ""
}

type SetFTPPasvY struct{}

func (c *SetFTPPasvY) Capacity() int {
	return 0
}

func (c *SetFTPPasvY) encode(fsm *FSM, template string, data []byte) string {
	i := fsm.Var("ftp_pasv_port").(int)
	return strconv.Itoa(i % 256)
}

func (c *SetFTPPasvY) decode(fsm *FSM, ciphertext string) string {
	ftp_pasv_port_x := fsm.Var("ftp_pasv_port").(int)
	ftp_pasv_port_y, _ := strconv.Atoi(ciphertext)

	fsm.SetVar("ftp_pasv_port", ftp_pasv_port_x*256+ftp_pasv_port_y)
	return ""
}

type SetDnsTransactionId struct{}

func (c *SetDnsTransactionId) capacity() int {
	return 0
}

func (c *SetDnsTransactionId) encode(fsm *FSM, template string, data []byte) string {
	var dns_transaction_id string
	if v := fsm.Var("dns_transaction_id"); v != nil {
		dns_transaction_id = v.(string)
	} else {
		dns_transaction_id = string([]rune{rune(rand.Intn(253) + 1), rune(rand.Intn(253) + 1)})
		fsm.SetVar("dns_transaction_id", dns_transaction_id)
	}
	return dns_transaction_id
}

func (c *SetDnsTransactionId) decode(fsm *FSM, ciphertext string) string {
	marionette_state.set_local("dns_transaction_id", ciphertext)
	return ""
}

type SetDnsDomain struct{}

func (c *SetDnsDomain) capacity() int {
	return 0
}

func (c *SetDnsDomain) encode(fsm *FSM, template string, data []byte) string {
	var dns_domain string
	if v := fsm.Var("dns_domain"); v != nil {
		dns_domain = v.(string)
	} else {
		available := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
		tlds := []string{"com", "net", "org"}

		buf := make([]rune, rand.Intn(60)+3+1)
		buf[0] = len(buf) - 1 // name length
		for i := 1; i < len(buf); i++ {
			buf[i] = available[rand.Intn(len(available))]
		}
		buf = append(buf, 3) // tld length
		buf = append(buf, tlds[rand.Intn(len(tlds))]...)

		dns_domain = string(buf)
		fsm.SetVar("dns_domain", dns_domain)
	}
	return dns_domain
}

func (c *SetDnsDomain) decode(fsm *FSM, ciphertext string) string {
	marionette_state.set_local("dns_domain", ctxt)
	return None
}

type SetDnsIp struct{}

func (c *SetDnsIp) capacity() int {
	return 0
}

func (c *SetDnsIp) encode(fsm *FSM, template string, data []byte) string {
	var dns_ip string
	if v := fsm.Var("dns_ip"); v != nil {
		dns_ip = v.(string)
	} else {
		dns_ip = string([]rune{rune(rand.Intn(253) + 1), rune(rand.Intn(253) + 1), rune(rand.Intn(253) + 1), rune(rand.Intn(253) + 1)})
		fsm.SetVar("dns_ip", dns_ip)
	}
	return dns_ip
}

func (c *SetDnsIp) decode(fsm *FSM, ciphertext string) string {
	marionette_state.set_local("dns_ip", ciphertext)
	return ""
}

var amazon_msg_lens = map[int]int{
	2049:  1,
	2052:  2,
	2054:  2,
	2057:  3,
	2058:  2,
	2059:  1,
	2065:  1,
	17429: 1,
	3098:  1,
	687:   3,
	2084:  1,
	42:    58,
	43:    107,
	9260:  1,
	11309: 1,
	11829: 1,
	9271:  1,
	6154:  1,
	64:    15,
	1094:  1,
	12376: 1,
	89:    1,
	10848: 1,
	5223:  1,
	69231: 1,
	7795:  1,
	2678:  1,
	8830:  1,
	29826: 1,
	16006: 10,
	8938:  1,
	17055: 2,
	87712: 1,
	23202: 1,
	7441:  1,
	17681: 1,
	12456: 1,
	41132: 1,
	25263: 6,
	689:   1,
	9916:  1,
	10101: 2,
	1730:  1,
	10948: 1,
	26826: 1,
	6357:  1,
	13021: 2,
	1246:  4,
	19683: 1,
	1765:  1,
	1767:  1,
	1768:  1,
	1769:  4,
	1770:  6,
	1771:  3,
	1772:  2,
	1773:  4,
	1774:  4,
	1775:  1,
	1776:  1,
	1779:  1,
	40696: 1,
	767:   1,
	17665: 1,
	27909: 1,
	12550: 1,
	5385:  1,
	16651: 1,
	5392:  1,
	26385: 1,
	12056: 1,
	41245: 2,
	13097: 1,
	15152: 1,
	310:   1,
	40759: 1,
	9528:  1,
	8000:  7,
	471:   1,
	15180: 1,
	14158: 3,
	37719: 2,
	1895:  1,
	31082: 1,
	19824: 1,
	30956: 1,
	18807: 1,
	11095: 1,
	37756: 2,
	746:   1,
	10475: 1,
	4332:  1,
	35730: 1,
	11667: 1,
	16788: 1,
	12182: 4,
	39663: 1,
	9126:  1,
	35760: 1,
	12735: 1,
	6594:  1,
	451:   15,
	19402: 1,
	463:   3,
	10193: 1,
	16853: 6,
	982:   1,
	15865: 1,
	2008:  2,
	476:   1,
	13655: 1,
	10213: 1,
	10737: 1,
	15858: 1,
	2035:  6,
	2039:  1,
	2041:  2,
}

/*

MIN_PTXT_LEN = fte.encoder.DfaEncoderObject._COVERTEXT_HEADER_LEN_CIPHERTTEXT + \
    fte.encrypter.Encrypter._CTXT_EXPANSION + 32
*/

/*

class AmazonMsgLensHandler(object):
    def __init__(self, regex, min_len = MIN_PTXT_LEN, msg_lens = amazon_msg_lens):
        self.regex_ = regex

        self.msg_lens_ = msg_lens
        self.random_lens_ = []
        for key in self.msg_lens_:
            self.random_lens_ += [key] * self.msg_lens_[key]

        self.min_len_ = min_len

        key = self.regex_ + str(self.min_len_)
        if not fte_cache_.get(key):
            dfa = regex2dfa.regex2dfa(self.regex_)
            encoder = fte.encoder.DfaEncoder(dfa, self.min_len_)
            fte_cache_[key] = (dfa, encoder)

        self.max_len_ = 2**18

        self.target_len_ = 0.0

*/

/*

   def capacity(self):

       self.target_len_ = random.choice(self.random_lens_)

       if self.target_len_ < self.min_len_:
           ptxt_len = 0.0

       elif self.target_len_ > self.max_len_:
           #We do this to prevent unranking really large slices
           # in practice this is probably bad since it unnaturally caps
           # our message sizes to whatever FTE can support
           ptxt_len = self.max_len_
           self.target_len_ = self.max_len_

       else:
           ptxt_len = self.target_len_ - fte.encoder.DfaEncoderObject._COVERTEXT_HEADER_LEN_CIPHERTTEXT
           ptxt_len -= fte.encrypter.Encrypter._CTXT_EXPANSION
           ptxt_len = int(ptxt_len * 8.0)-1

       return ptxt_len
*/

/*

   def encode(self, marionette_state, template, to_embed):
       ctxt = ''

       if self.target_len_ < self.min_len_ or self.target_len_ > self.max_len_:
           key = self.regex_ + str(self.target_len_)
           if not regex_cache_.get(key):
               dfa = regex2dfa.regex2dfa(self.regex_)
               cdfa_obj = fte.cDFA.DFA(dfa, self.target_len_)
               encoder = fte.dfa.DFA(cdfa_obj, self.target_len_)
               regex_cache_[key] = (dfa, encoder)

           (dfa, encoder) = regex_cache_[key]

           to_unrank = random.randint(0, encoder.getNumWordsInSlice(self.target_len_))
           ctxt = encoder.unrank(to_unrank)

       else:
           key = self.regex_ + str(self.min_len_)
           (dfa, encoder) = fte_cache_[key]

           ctxt = encoder.encode(to_embed)

           if len(ctxt) != self.target_len_:
               raise Exception("Could not find ctxt of len %d (%d)" %
                   (self.target_len_,len(ctxt)))

       return ctxt
*/

/*

   def decode(self, marionette_state, ctxt):
       ptxt = None

       ctxt_len = len(ctxt)

       if ctxt_len >= self.min_len_:
           key = self.regex_ + str(self.min_len_)
           (dfa, encoder) = fte_cache_[key]

           try:
               retval = encoder.decode(ctxt)
               ptxt = retval[0]
           except Exception as e:
               pass

       return ptxt
*/

type tgConfig struct {
	Grammar  string
	Handlers []*tgHandler
}

type tgHandler struct {
	Name    string
	Crypter tgCrypter
}

type tgCrypter interface {
	Capacity() int
	Encrypt(fsm *FSM, template string, data []byte) string
}

var tgConfigs = map[string]*tgConfig{
	"http_request_keep_alive": &tgConfig{
		Grammar: "http_request_keep_alive",
		Handlers: []*tgHandler{
			{Name: "URL", Crypter: RankerCrypter(`[a-zA-Z0-9\?\-\.\&]+`, 2048)},
		},
	},

	"http_response_keep_alive": &tgConfig{
		Grammar: "http_response_keep_alive",
		Handlers: []*tgHandler{
			{Name: "HTTP-RESPONSE-BODY", Crypter: FteCrypter(".+", 128, false)},
			{Name: "CONTENT-LENGTH", Crypter: HttpContentLengthCrypter()},
		},
	},

	"http_request_close": &tgConfig{
		Grammar:  "http_request_close",
		Handlers: []*tgHandler{{Name: "URL", Crypter: RankerCrypter(`[a-zA-Z0-9\?\-\.\&]+`, 2048)}},
	},

	"http_response_close": &tgConfig{
		Grammar: "http_response_close",
		Handlers: []*tgHandler{
			{Name: "HTTP-RESPONSE-BODY", Crypter: FteCrypter(`.+`, 128, false)},
			{Name: "CONTENT-LENGTH", Crypter: HttpContentLengthCrypter()},
		},
	},

	"pop3_message_response": &tgConfig{
		Grammar: "pop3_message_response",
		Handlers: []*tgHandler{
			{Name: "POP3-RESPONSE-BODY", Crypter: RankerCrypter(`[a-zA-Z0-9]+`, 2048)},
			{Name: "CONTENT-LENGTH", Crypter: Pop3ContentLengthCrypter()},
		},
	},

	"pop3_password": &tgConfig{
		Grammar:  "pop3_password",
		Handlers: []*tgHandler{{Name: "PASSWORD", Crypter: RankerCrypter(`[a-zA-Z0-9]+`, 256)}},
	},

	"http_request_keep_alive_with_msg_lens": &tgConfig{
		Grammar:  "http_request_keep_alive",
		Handlers: []*tgHandler{{Name: "URL", Crypter: FteCrypter(`[a-zA-Z0-9\?\-\.\&]+`, 2048, true)}},
	},

	"http_response_keep_alive_with_msg_lens": &tgConfig{
		Grammar:      "http_response_keep_alive",
		HandlerOrder: []string{"HTTP-RESPONSE-BODY", "CONTENT-LENGTH"},
		Handlers: []*tgHandler{
			{Name: "HTTP-RESPONSE-BODY", Crypter: FteCrypter(`.+`, 2048, true)},
			{Name: "CONTENT-LENGTH", Crypter: HttpContentLengthCrypter()},
		},
	},

	"http_amazon_request": &tgConfig{
		Grammar:  "http_request_keep_alive",
		Handlers: []*tgHandler{{Name: "URL", Crypter: RankerCrypter(`[a-zA-Z0-9\?\-\.\&]+`, 2048)}},
	},

	"http_amazon_response": &tgConfig{
		Grammar: "http_response_keep_alive",
		Handlers: []*tgHandler{
			{Name: "HTTP-RESPONSE-BODY", Crypter: AmazonMsgLensHandler(`.+`)},
			{Name: "CONTENT-LENGTH", Crypter: HttpContentLengthCrypter()},
		},
	},

	"ftp_entering_passive": &tgConfig{
		Grammar: "ftp_entering_passive",
		Handlers: []*tgHandler{
			{Name: "FTP_PASV_PORT_X", Crypter: SetFTPPasvX()},
			{Name: "FTP_PASV_PORT_Y", Crypter: SetFTPPasvY()},
		},
	},

	"dns_request": &tgConfig{
		Grammar: "dns_request",
		Handlers: []*tgHandler{
			{Name: "DNS_TRANSACTION_ID", Crypter: SetDnsTransactionId()},
			{Name: "DNS_DOMAIN", Crypter: SetDnsDomain()},
		},
	},

	"dns_response": &tgConfig{
		Grammar: "dns_response",
		Handlers: []*tgHandler{
			{Name: "DNS_TRANSACTION_ID", Crypter: SetDnsTransactionId()},
			{Name: "DNS_DOMAIN", Crypter: SetDnsDomain()},
			{Name: "DNS_IP", Crypter: SetDnsIp()},
		},
	},
}

/*

# grammars


def parser(grammar, msg):
    if grammar.startswith(
            "http_response") or grammar == "http_amazon_response":
        return http_response_parser(msg)
    elif grammar.startswith("http_request") or grammar == "http_amazon_request":
        return http_request_parser(msg)
    elif grammar.startswith("pop3_message_response"):
        return pop3_parser(msg)
    elif grammar.startswith("pop3_password"):
        return pop3_password_parser(msg)
    elif grammar.startswith("ftp_entering_passive"):
        return ftp_entering_passive_parser(msg)
    elif grammar.startswith("dns_request"):
        return dns_request_parser(msg)
    elif grammar.startswith("dns_response"):
        return dns_response_parser(msg)


*/

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

/*

def get_http_header(header_name, msg):
    retval = None

    message_lines = msg.split("\r\n")
    for line in message_lines[1:-2]:
        line_compontents = line.partition(": ")
        if line_compontents[0] == header_name:
            retval = line_compontents[-1]
            break

    return retval
*/

/*


def http_request_parser(msg):
    if not msg.startswith("GET"):
        return None

    retval = {}

    if msg.startswith("GET http"):
        retval["URL"] = '/'.join(msg.split('\r\n')[0][:-9].split('/')[3:])
    else:
        retval["URL"] = '/'.join(msg.split('\r\n')[0][:-9].split('/')[1:])

    if not msg.endswith("\r\n\r\n"):
        retval = None

    return retval

*/

/*

def http_response_parser(msg):
    if not msg.startswith("HTTP"):
        return None

    retval = {}

    retval["CONTENT-LENGTH"] = int(get_http_header("Content-Length", msg))
    retval["COOKIE"] = get_http_header("Cookie", msg)
    try:
        retval["HTTP-RESPONSE-BODY"] = msg.split("\r\n\r\n")[1]
    except:
        retval["HTTP-RESPONSE-BODY"] = ''

    if retval["CONTENT-LENGTH"] != len(retval["HTTP-RESPONSE-BODY"]):
        retval = None

    return retval

*/

/*

def pop3_parser(msg):
    retval = {}

    try:
        retval["POP3-RESPONSE-BODY"] = msg.split('\n\n')[1]
        assert retval["POP3-RESPONSE-BODY"].endswith('\n.\n')
        retval["POP3-RESPONSE-BODY"] = retval["POP3-RESPONSE-BODY"][:-3]
        retval["CONTENT-LENGTH"] = len(retval["POP3-RESPONSE-BODY"])
    except Exception as e:
        pass
        retval = {}

    return retval

*/

/*

def pop3_password_parser(msg):
    retval = {}

    try:
        assert msg.endswith('\n')
        retval["PASSWORD"] = msg[5:-1]
    except Exception as e:
        retval = {}

    return retval
*/

/*

def ftp_entering_passive_parser(msg):
    retval = {}

    try:
        assert msg.startswith("227 Entering Passive Mode (")
        assert msg.endswith(").\n")
        bits = msg.split(',')
        retval['FTP_PASV_PORT_X'] = int(bits[4])
        retval['FTP_PASV_PORT_Y'] = int(bits[5][:-3])
    except Exception as e:
        retval = {}

    return retval

*/

/*

def validate_dns_domain(msg, dns_response=False):
    if dns_response:
        expected_splits = 3
        split1_msg = '\x81\x80\x00\x01\x00\x01\x00\x00\x00\x00'
    else:
        expected_splits = 2
        split1_msg = '\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00'

    tmp_domain_split1 = msg.split(split1_msg)
    if len(tmp_domain_split1) != 2:
        return None
    tmp_domain_split2 = tmp_domain_split1[1].split('\x00\x01\x00\x01')
    if len(tmp_domain_split2) != expected_splits:
        return None
    tmp_domain = tmp_domain_split2[0]
    # Check for valid prepended length
    # Remove trailing tld prepended length (1), tld (3) and trailing null (1) = 5
    if ord(tmp_domain[0]) != len(tmp_domain[1:-5]):
        return None
    if ord(tmp_domain[-5]) != 3:
        return None
    # Check for valid TLD
    if not re.search("(com|net|org)\x00$", tmp_domain):
        return None
    # Check for valid domain characters
    if not re.match("^[\w\d]+$", tmp_domain[1:-5]):
        return None

    return tmp_domain

*/

/*

def validate_dns_ip(msg):
    tmp_ip_split = msg.split('\x00\x01\x00\x01\xc0\x0c\x00\x01\x00\x01\x00\x00\x00\x02\x00\x04')
    if len(tmp_ip_split) != 2:
        return None
    tmp_ip = tmp_ip_split[1]
    if len(tmp_ip) != 4:
        return None

    return tmp_ip

*/

/*

def dns_request_parser(msg):
    retval = {}
    if '\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00' not in msg:
        return retval

    try:
        # Nothing to validate for Transaction ID
        retval["DNS_TRANSACTION_ID"] = msg[:2]

        tmp_domain = validate_dns_domain(msg)
        if not tmp_domain:
            raise Exception("Bad DNS Domain")
        retval["DNS_DOMAIN"] = tmp_domain

    except Exception as e:
        retval = {}

    return retval

*/

/*

def dns_response_parser(msg):
    retval = {}
    if '\x81\x80\x00\x01\x00\x01\x00\x00\x00\x00' not in msg:
        return retval

    try:
        # Nothing to validate for Transaction ID
        retval["DNS_TRANSACTION_ID"] = msg[:2]

        tmp_domain = validate_dns_domain(msg, dns_response=True)
        if not tmp_domain:
            raise Exception("Bad DNS Domain")
        retval["DNS_DOMAIN"] = tmp_domain

        tmp_ip = validate_dns_ip(msg)
        if not tmp_ip:
            raise Exception("Bad DNS IP")
        retval["DNS_IP"] = tmp_ip

    except Exception as e:
        retval = {}

    return retval

*/
