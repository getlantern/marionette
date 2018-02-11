package tg

import (
	"strconv"
	"strings"

	"github.com/redjack/marionette"
)

type SetFTPPasvXCipher struct{}

func NewSetFTPPasvXCipher() *SetFTPPasvXCipher {
	return &SetFTPPasvXCipher{}
}

func (c *SetFTPPasvXCipher) Key() string {
	return "FTP_PASV_PORT_X"
}

func (c *SetFTPPasvXCipher) Capacity() int {
	return 0
}

func (c *SetFTPPasvXCipher) Encrypt(fsm marionette.FSM, template string, plaintext []byte) (ciphertext []byte, err error) {
	i := fsm.Var("ftp_pasv_port").(int)
	return []byte(strconv.Itoa(i / 256)), nil
}

func (c *SetFTPPasvXCipher) Decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	i, _ := strconv.Atoi(string(ciphertext))
	fsm.SetVar("ftp_pasv_port_x", i)
	return nil, nil
}

type SetFTPPasvYCipher struct{}

func NewSetFTPPasvYCipher() *SetFTPPasvYCipher {
	return &SetFTPPasvYCipher{}
}

func (c *SetFTPPasvYCipher) Key() string {
	return "FTP_PASV_PORT_Y"
}

func (c *SetFTPPasvYCipher) Capacity() int {
	return 0
}

func (c *SetFTPPasvYCipher) Encrypt(fsm marionette.FSM, template string, plaintext []byte) (ciphertext []byte, err error) {
	i := fsm.Var("ftp_pasv_port").(int)
	return []byte(strconv.Itoa(i % 256)), nil
}

func (c *SetFTPPasvYCipher) Decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	ftp_pasv_port_x := fsm.Var("ftp_pasv_port").(int)
	ftp_pasv_port_y, _ := strconv.Atoi(string(ciphertext))

	fsm.SetVar("ftp_pasv_port", ftp_pasv_port_x*256+ftp_pasv_port_y)
	return nil, nil
}

func parseFTPEnteringPassive(msg string) map[string]string {
	if !strings.HasPrefix(msg, "227 Entering Passive Mode (") || !strings.HasSuffix(msg, ").\n") {
		return nil
	}

	a := strings.Split(msg, ",")
	if len(a) < 6 {
		return nil
	}

	return map[string]string{
		"FTP_PASV_PORT_X": a[4],
		"FTP_PASV_PORT_Y": strings.TrimSuffix(a[5], ").\n"),
	}
}
