package d2dt1

// Lots of unknowns for now
type MaterialFlags struct {
	Animated bool // Is only lava animated? Might just be a lava flag
}

func NewMaterialFlags(data uint16) MaterialFlags {
	return MaterialFlags{
		Animated: data&0x0100 == 0x0100,
	}
}
