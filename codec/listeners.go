package codec

type packetHandler func(*Client, Packet) error

type listeners map[string][]packetHandler

func (l listeners) add(name string, handler packetHandler) {
	l[name] = append(l[name], handler)
}

func (l listeners) dispatch(c *Client, packet Packet) error {
	for _, handler := range l[packet.Name()] {
		if err := handler(c, packet); err != nil {
			return err
		}
	}

	return nil
}
