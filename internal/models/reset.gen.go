package models

func (v *InsertData) Reset() {

	v.Key = ""

	v.Value = ""

	v.UserId = ""

	if v.MyCustom != nil {

		*v.MyCustom = ""

	}

	v.Gen = 0

	v.BoolVal = false

	if v.StarBoolVar != nil {

		*v.StarBoolVar = false

	}

	v.Sll = v.Sll[:0]

	clear(v.m)

	v.child.Reset()

	v.MyStrrruct.Reset()

	if v.MyStructWithStar != nil {
		v.MyStructWithStar.Reset()
	}

}

func (v *ShortenBatchReq) Reset() {

	v.CorrelationId = ""

	v.OriginalUrl = ""

}
