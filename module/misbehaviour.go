package module

func (*Misbehaviour) ClientType() string {
	return Parlia
}

func (h *Misbehaviour) GetClientID() string {
	return h.ClientId
}

func (h *Misbehaviour) ValidateBasic() error {
	if err := h.Header_1.ValidateBasic(); err != nil {
		return err
	}
	if err := h.Header_2.ValidateBasic(); err != nil {
		return err
	}
	return nil
}
