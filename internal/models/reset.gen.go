package models

func (v *InsertData) Reset() {
	v.Key = ""

	v.Value = ""

	v.UserId = ""
}

func (v *ShortenBatchReq) Reset() {
	v.CorrelationId = ""

	v.OriginalUrl = ""
}
