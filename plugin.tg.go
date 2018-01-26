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

func get_grammar_capacity(grammar) {
    var n int
    for _, h := range tgConfigs[grammar].Handlers {
        retval += h.capacity()
    }
    return retval
}


func do_embed(template, handler_key, value string) {
	return strings.Replace(template, "%%" + handler_key + "%%", value, -1)
}

func do_unembed(grammar, ciphertext, handler_key string) string {
    parse_tree := parser(grammar, ciphertext)
    return parse_tree[handler_key]
}

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
	return do_embed(template, handler, value_to_embed)
}

func execute_handler_receiver(fsm *FSM, grammar, handler_key, ciphertext string) string {
    handler_key_value = do_unembed(grammar, ciphertext, handler_key)

    to_execute := conf[grammar].Handler(handler_key)
    return to_execute.decode(marionette_state, handler_key_value)
}

var regex_cache_ = {}
var fte_cache_ = {}

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


var amazon_msg_lens []int

func init() { 
	for _, item := range []struct {
		n int
		weight int
	}{
		{n: 2049:  weight:1},
		{n: 2052:  weight:2},
		{n: 2054:  weight:2},
		{n: 2057:  weight:3},
		{n: 2058:  weight:2},
		{n: 2059:  weight:1},
		{n: 2065:  weight:1},
		{n: 17429: weight:1},
		{n: 3098:  weight:1},
		{n: 687:   weight:3},
		{n: 2084:  weight:1},
		{n: 42:    weight:58},
		{n: 43:    weight:107},
		{n: 9260:  weight:1},
		{n: 11309: weight:1},
		{n: 11829: weight:1},
		{n: 9271:  weight:1},
		{n: 6154:  weight:1},
		{n: 64:    weight:15},
		{n: 1094:  weight:1},
		{n: 12376: weight:1},
		{n: 89:    weight:1},
		{n: 10848: weight:1},
		{n: 5223:  weight:1},
		{n: 69231: weight:1},
		{n: 7795:  weight:1},
		{n: 2678:  weight:1},
		{n: 8830:  weight:1},
		{n: 29826: weight:1},
		{n: 16006: weight:10},
		{n: 8938:  weight:1},
		{n: 17055: weight:2},
		{n: 87712: weight:1},
		{n: 23202: weight:1},
		{n: 7441:  weight:1},
		{n: 17681: weight:1},
		{n: 12456: weight:1},
		{n: 41132: weight:1},
		{n: 25263: weight:6},
		{n: 689:   weight:1},
		{n: 9916:  weight:1},
		{n: 10101: weight:2},
		{n: 1730:  weight:1},
		{n: 10948: weight:1},
		{n: 26826: weight:1},
		{n: 6357:  weight:1},
		{n: 13021: weight:2},
		{n: 1246:  weight:4},
		{n: 19683: weight:1},
		{n: 1765:  weight:1},
		{n: 1767:  weight:1},
		{n: 1768:  weight:1},
		{n: 1769:  weight:4},
		{n: 1770:  weight:6},
		{n: 1771:  weight:3},
		{n: 1772:  weight:2},
		{n: 1773:  weight:4},
		{n: 1774:  weight:4},
		{n: 1775:  weight:1},
		{n: 1776:  weight:1},
		{n: 1779:  weight:1},
		{n: 40696: weight:1},
		{n: 767:   weight:1},
		{n: 17665: weight:1},
		{n: 27909: weight:1},
		{n: 12550: weight:1},
		{n: 5385:  weight:1},
		{n: 16651: weight:1},
		{n: 5392:  weight:1},
		{n: 26385: weight:1},
		{n: 12056: weight:1},
		{n: 41245: weight:2},
		{n: 13097: weight:1},
		{n: 15152: weight:1},
		{n: 310:   weight:1},
		{n: 40759: weight:1},
		{n: 9528:  weight:1},
		{n: 8000:  weight:7},
		{n: 471:   weight:1},
		{n: 15180: weight:1},
		{n: 14158: weight:3},
		{n: 37719: weight:2},
		{n: 1895:  weight:1},
		{n: 31082: weight:1},
		{n: 19824: weight:1},
		{n: 30956: weight:1},
		{n: 18807: weight:1},
		{n: 11095: weight:1},
		{n: 37756: weight:2},
		{n: 746:   weight:1},
		{n: 10475: weight:1},
		{n: 4332:  weight:1},
		{n: 35730: weight:1},
		{n: 11667: weight:1},
		{n: 16788: weight:1},
		{n: 12182: weight:4},
		{n: 39663: weight:1},
		{n: 9126:  weight:1},
		{n: 35760: weight:1},
		{n: 12735: weight:1},
		{n: 6594:  weight:1},
		{n: 451:   weight:15},
		{n: 19402: weight:1},
		{n: 463:   weight:3},
		{n: 10193: weight:1},
		{n: 16853: weight:6},
		{n: 982:   weight:1},
		{n: 15865: weight:1},
		{n: 2008:  weight:2},
		{n: 476:   weight:1},
		{n: 13655: weight:1},
		{n: 10213: weight:1},
		{n: 10737: weight:1},
		{n: 15858: weight:1},
		{n: 2035:  weight:6},
		{n: 2039:  weight:1},
		{n: 2041:  weight:2},
	} {
		for i := 0; i<item.weight; i++ {
			amazon_msg_lens = append(amazon_msg_lens, item.n)
		}
	}
}

cont MIN_PTXT_LEN = fte.encoder.DfaEncoderObject._COVERTEXT_HEADER_LEN_CIPHERTTEXT + fte.encrypter.Encrypter._CTXT_EXPANSION + 32

type AmazonMsgLensHandler struct {}

func NewAmazonMsgLensHandler(regex string) *AmazonMsgLensHandler {
        // key = self.regex_ + str(self.min_len_)
        // if not fte_cache_.get(key):
        //     dfa = regex2dfa.regex2dfa(self.regex_)
        //     encoder = fte.encoder.DfaEncoder(dfa, self.min_len_)
        //     fte_cache_[key] = (dfa, encoder)

	return &AmazonMsgLensHandler {
		min_len: MIN_PTXT_LEN,
		max_len_: 1 << 18,
		target_len_: 0.0,
	    regex_: regex,
	}
}

func    (h *AmazonMsgLensHandler) capacity() int {
       h.target_len_ = amazon_msg_lens[rand.Intn(amazon_msg_lens)]
       if h.target_len_ < h.min_len_ {
           return 0
       } else if  h.target_len_ > h.max_len_ {
           // We do this to prevent unranking really large slices
           // in practice this is probably bad since it unnaturally caps
           // our message sizes to whatever FTE can support
           h.target_len_ = h.max_len_
           return h.max_len_
	   } 
       n := h.target_len_ - fte.encoder.DfaEncoderObject._COVERTEXT_HEADER_LEN_CIPHERTTEXT
       n -= fte.encrypter.Encrypter._CTXT_EXPANSION
       // n = int(ptxt_len * 8.0)-1
       return n
}

   func    (h *AmazonMsgLensHandler)  encode( marionette_state, template, to_embed) {
       if h.target_len_ < h.min_len_ || h.target_len_ > h.max_len_ {
           key := h.regex_ + str(h.target_len_)
           if not regex_cache_.get(key) {
               dfa := regex2dfa.regex2dfa(h.regex_)
               cdfa_obj := fte.cDFA.DFA(dfa, h.target_len_)
               encoder := fte.dfa.DFA(cdfa_obj, h.target_len_)
               regex_cache_[key] = encoder
			}
           encoder := regex_cache_[key]

           to_unrank := random.Intn(encoder.getNumWordsInSlice(h.target_len_))
           return encoder.unrank(to_unrank)
		}
       
           key := h.regex_ + str(h.min_len_)
           encoder := fte_cache_[key]

           ciphertext := encoder.encode(to_embed)

           if len(ciphertext) != h.target_len_ {
               return fmt.Errorf("Could not find ciphertext of len %d (%d)" , h.target_len_, len(ciphertext))
           }

       return ciphertext
}


   func    (h *AmazonMsgLensHandler)  decode(self, marionette_state, ctxt) {
       if len(ctxt) < self.min_len_ {
       	return ""
       }
       key := self.regex_ + str(self.min_len_)
        encoder = fte_cache_[key]

       plaintext := encoder.decode(ctxt)
       return plaintext
}

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

func  parse(grammar, msg) {
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
    return ""
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
    lines := msg.split("\r\n")
    for _, line := range lines[1:len(lines)-2] {        
        if a := strings.Split(line, ": ", 2); a[0] == header_name {
            if len(a) > 1 {
            	return a[1]
        	}
        	return ""
        }
	}
    return ""
}


func http_request_parser(msg string) map[string]interface{} {
    if !strings.HasPrefix( msg.startswith("GET") ) {
        return nil
    } else if !strings.HasSuffix(msg, "\r\n\r\n") {
        return nil
    }


	lines := lineBreakRegex.Split(msg)
	segments := strings.Split(lines[0][:len(lines[0])-9], "/")

    if strings.HasPrefix(msg, "GET http") {
        return map[string]interface{}{"URL": strings.Join(segments[3:], "/")}
    } 
    return map[string]interface{}{"URL": strings.Join(segments[1:], "/")}
}

var lineBreakRegex = regexp.MustCompile(`\r\n`)


func http_response_parser(msg)map[string]interface{} {
    if !strings.HasPrefix(msg, "HTTP") {
        return nil
    }

    m :=  make(map[string]interface{})
    m["CONTENT-LENGTH"], _ = strconv.Atoi(get_http_header("Content-Length", msg))
    m["COOKIE"] = get_http_header("Cookie", msg)
    if a := strings.Split(msg< "\r\n\r\n"); len(a) > 1 {
        m["HTTP-RESPONSE-BODY"] = a[1]
    } else {
        m["HTTP-RESPONSE-BODY"] = ""
       }

    if m["CONTENT-LENGTH"] != len(m["HTTP-RESPONSE-BODY"]) {
        return nil
    }

    return m
}



func pop3_parser(msg) map[string]interface{} {
    a := strings.Split(msg, "\n\n"); 
    if len(a) < 2 {
    	return make(map[string]interface{})
    }
    
    body := a[1]
	if !strings.HasSuffix(body, "\n.\n") {
    	return fmt.Errorf("invalid POP3-RESPONSE-BODY")
    }
    body = strings.TrimSuffix(body ,"\n.\n")
    

    return map[string]interface{}{
		"POP3-RESPONSE-BODY": body,
		"CONTENT-LENGTH": len(body),
    }
}

func pop3_password_parser(msg) map[string]interface{} {
    if  !strings.HasSuffix(msg, "\n") {
    	return nil
    }
    return map[string]interface{}{
    	"PASSWORD": msg[5:len(msg)-1],
    }
}

func ftp_entering_passive_parser(msg) map[string]interface{} {
    if !strings.HasPrefix(msg, "227 Entering Passive Mode (") || !strings.HasSuffix(msg, ").\n") {
    	return make(map[string]interface{})
    }

	a := msg.split(',')
	if len(a) < 6 {
		return make(map[string]interface{})
	}

	return map[string]interface{}{
        "FTP_PASV_PORT_X": int(a[4]),
        "FTP_PASV_PORT_Y": int(a[5][:len(a[5])-3]),
	}
}


func validate_dns_domain(msg string, dns_response bool) string {
    delim, splitN :=  "\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00", 2
    if dns_response {
        delim, splitN =  "\x81\x80\x00\x01\x00\x01\x00\x00\x00\x00", 3
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
    } else if int(tmp_domain[-5]) != 3 {
        return ""
    }

    // Check for valid TLD
    if !strings.HasSuffix(tmp_domain, "com\x00") && !strings.HasSuffix(tmp_domain, "net\x00") && !strings.HasSuffix(tmp_domain, "org\x00") {
    	return ""
    }
    
    // Check for valid domain characters
    if ! domainRegex.MatchString(tmp_domain[1:len(tmp_domain)-5]) {
        return ""
    }

    return tmp_domain
}

var domainRegex = regexp.MustCompile(`^[\w\d]+$`)


func validate_dns_ip(msg string)  string {
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

func dns_request_parser(msg string)  map[string]interface{} {
    if !strings.Contains(msg, "\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00") {
        return make(map[string]interface{})
    }

    tmp_domain := validate_dns_domain(msg, false)
    if tmp_domain == "" {
        return make(map[string]interface{})
    }

    return map[string]interface{}{
        "DNS_TRANSACTION_ID": msg[:2],
    	"DNS_DOMAIN": tmp_domain,
    }
}

func dns_response_parser(msg string)   map[string]interface{} {
    if !strings.Contains(msgm "\x81\x80\x00\x01\x00\x01\x00\x00\x00\x00") {
        return make(map[string]interface{})
    }

    tmp_domain := validate_dns_domain(msg, True)
    if tmp_domain == "" {
        return return make(map[string]interface{})
    }

    tmp_ip := validate_dns_ip(msg)
    if tmp_ip == "" {
    	return return make(map[string]interface{})
    }

    return map[string]interface{}{
        "DNS_TRANSACTION_ID": msg[:2],
        "DNS_DOMAIN": tmp_domain,
        "DNS_IP": tmp_ip,
    }
}