package d2ds1

type WallRecord struct {
	Type        byte
	Zero        byte
	Prop1       byte
	Sequence    byte
	Unknown1    byte
	Style       byte
	Unknown2    byte
	Hidden      bool
	RandomIndex byte
	YAdjust     int
}
