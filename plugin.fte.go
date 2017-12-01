package marionette

import (
	"errors"
	"math"

	"github.com/redjack/marionette/fte"
)

const MaxCellLengthInBits = 2097152 // (2 ^ 18) * 8

// FTESendPlugin send data to a connection.
func FTESendPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	return fteSendPlugin(fsm, args, true)
}

// FTESendAsyncPlugin send data to a connection without blocking.
func FTESendAsyncPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	return fteSendPlugin(fsm, args, false)
}

func fteSendPlugin(fsm *FSM, args []interface{}, blocking bool) (success bool, err error) {
	if len(args) < 2 {
		return false, errors.New("fte.send: not enough arguments")
	}

	regex, ok := args[0].(string)
	if !ok {
		return false, errors.New("fte.send: invalid regex argument type")
	}
	msgLen, ok := args[1].(int)
	if !ok {
		return false, errors.New("fte.send: invalid msg_len argument type")
	}

	// Find random stream id with data.
	streamID := fsm.enc.chooseStreamID()
	if streamID == 0 && !blocking {
		return false, nil
	}

	fteEncoder := fsm.fteEncoder(regex, msgLen)

	bufferBitN := len(fsm.enc.Peek(streamID)) * 8
	minCellByteN := int(math.Floor(float64(fteEncoder.Capacity())/8.0)) - fte.COVERTEXT_HEADER_LEN_CIPHERTTEXT - fte.CTXT_EXPANSION
	minCellBitN := minCellByteN * 8

	cellHeaderBitN := PAYLOAD_HEADER_SIZE_IN_BITS
	cellBitN := minCellBitN
	if bufferBitN > cellBitN {
		cellBitN = bufferBitN
	}
	cellBitN += cellHeaderBitN
	if cellBitN > MaxCellLengthInBits {
		cellBitN = MaxCellLengthInBits
	}

	cell, err := fsm.enc.Pop(fsm.ModelUUID(), fsm.ModelInstanceID, cellBitN)
	if err != nil {
		return false, err
	}

	plaintext, err := cell.MarshalBinary()
	if err != nil {
		return false, err
	}

	ciphertext, err := fteEncoder.Encode(plaintext)
	if err != nil {
		return false, err
	}

	if _, err := fsm.conn.Write(ciphertext); err != nil {
		return false, err
	}
	return true, nil
}

// FTERecvPlugin receives data from a connection.
func FTERecvPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	return fteRecvPlugin(fsm, args, true)
}

// FTERecvAsyncPlugin receives data from a connection without blocking.
func FTERecvAsyncPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	return fteRecvPlugin(fsm, args, false)
}

func fteRecvPlugin(fsm *FSM, args []interface{}, blocking bool) (success bool, err error) {
	/*
	   retval = False
	   regex = input_args[0]
	   msg_len = int(input_args[1])

	   fteObj = marionette_state.get_fte_obj(regex, msg_len)

	   try:
	       ctxt = channel.recv()
	       if len(ctxt) > 0:
	           [ptxt, remainder] = fteObj.decode(ctxt)

	           cell_obj = marionette_tg.record_layer.unserialize(ptxt)
	           assert cell_obj.get_model_uuid() == marionette_state.get_local(
	               "model_uuid")

	           marionette_state.set_local(
	               "model_instance_id", cell_obj.get_model_instance_id())

	           if marionette_state.get_local("model_instance_id"):
	               if cell_obj.get_stream_id() > 0:
	                   marionette_state.get_global(
	                       "multiplexer_incoming").push(ptxt)
	               retval = True
	   except fte.encrypter.RecoverableDecryptionError as e:
	       retval = False
	   except Exception as e:
	       if len(ctxt)>0:
	           channel.rollback()
	       raise e

	   if retval:
	       if len(remainder) > 0:
	           channel.rollback(len(remainder))
	   else:
	       if len(ctxt)>0:
	           channel.rollback()

	   return retval
	*/
	panic("TODO")
}
