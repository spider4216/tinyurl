package models

func (v *InsertData) Reset() {
	if v == nil {
		return
	}

	v.Key = ""

	v.Value = ""

	v.UserId = ""
}

func (v *ShortenReq) Reset() {
	if v == nil {
		return
	}

	v.Url = ""
}
