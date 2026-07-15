package models

func (v *InsertData) Reset() {
	v.Key = ""

	v.Value = ""

	v.UserId = ""
}

func (v *ShortenReq) Reset() {
	v.Url = ""
}
