package marionette

const (
	PAYLOAD_HEADER_SIZE_IN_BITS  = 200
	PAYLOAD_HEADER_SIZE_IN_BYTES = PAYLOAD_HEADER_SIZE_IN_BITS / 8
)

const (
	NORMAL        = 0x1
	END_OF_STREAM = 0x2
	NEGOTIATE     = 0x3
)

type Cell struct {
	cell_type         int
	payload           []byte
	payload_length    int
	sequence_id       int
	cell_length       int
	stream_id         int
	model_uuid        int
	model_instance_id int
}

func NewCell(model_uuid, model_instance_id, stream_id, seq_id, length, cell_type int) *Cell {
	assert(stream_id != 0)
	return &Cell{
		cell_type:         cell_type,
		sequence_id:       seq_id,
		cell_length:       length,
		stream_id:         stream_id,
		model_uuid:        model_uuid,
		model_instance_id: model_instance_id,
	}
}

func (c *Cell) Compare(other *Cell) int {
	self_id := c.sequence_id
	other_id := other.sequence_id
	if self_id < other_id {
		return -1
	} else if self_id > other_id {
		return 1
	}
	return 0
}

func (c *Cell) Equal(other *Cell) bool {
	return
	bytes.Equal(c.payload, other.payload) &&
		c.stream_id == other.stream_id &&
		c.model_uuid == other.model_uuid &&
		c.model_instance_id == other.model_instance_id &&
		c.seq_id == other.seq_id
}

func (c *Cell) String() string {
	return serialize(c, c.cell_length_)
}

func serialize(cell_obj *Cell, pad_to int) []byte {
	/*
	   retval = ''

	   stream_id = cell_obj.get_stream_id()
	   model_uuid = cell_obj.get_model_uuid()
	   model_instance_id = cell_obj.get_model_instance_id()
	   seq_id = cell_obj.get_seq_id()
	   payload = cell_obj.get_payload()
	   padding = '\x00' * (
	       (pad_to / 8) - len(payload) - PAYLOAD_HEADER_SIZE_IN_BYTES)
	   cell_type = cell_obj.get_cell_type()

	   bytes_cell_len = pad_to_bytes(4, long_to_bytes(
	       PAYLOAD_HEADER_SIZE_IN_BYTES + len(payload) + len(padding)))
	   bytes_payload_len = pad_to_bytes(4, long_to_bytes(len(payload)))
	   bytes_model_uuid = pad_to_bytes(4, long_to_bytes(model_uuid))
	   bytes_model_instance_id = pad_to_bytes(4, long_to_bytes(model_instance_id))
	   bytes_stream_id = pad_to_bytes(4, long_to_bytes(stream_id))
	   bytes_seq_id = pad_to_bytes(4, long_to_bytes(seq_id))
	   bytes_cell_type = pad_to_bytes(1, long_to_bytes(cell_type))

	   retval += bytes_cell_len
	   retval += bytes_payload_len
	   retval += bytes_model_uuid
	   retval += bytes_model_instance_id
	   retval += bytes_stream_id
	   retval += bytes_seq_id
	   retval += bytes_cell_type
	   retval += payload
	   retval += padding

	   assert (PAYLOAD_HEADER_SIZE_IN_BYTES + len(payload) + len(padding)
	           ) == len(retval)

	   return retval
	*/
	panic("TODO")
}

func unserialize(data []byte) *Cell {
	/*
	   cell_len = bytes_to_long(cell_str[:4])
	   payload_len = bytes_to_long(cell_str[4:8])
	   model_uuid = bytes_to_long(cell_str[8:12])
	   model_instance_id = bytes_to_long(cell_str[12:16])
	   stream_id = bytes_to_long(cell_str[16:20])
	   seq_id = bytes_to_long(cell_str[20:24])
	   cell_type = bytes_to_long(cell_str[24:25])

	   if cell_len != len(cell_str):
	       raise UnserializeException()

	   payload = cell_str[PAYLOAD_HEADER_SIZE_IN_BYTES:
	                      PAYLOAD_HEADER_SIZE_IN_BYTES + payload_len]

	   retval = Cell(
	       model_uuid,
	       model_instance_id,
	       stream_id,
	       seq_id,
	       payload_len,
	       cell_type)
	   retval.set_payload(payload)

	   return retval
	*/
	panic("TODO")
}
