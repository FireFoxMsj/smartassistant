package attribute

type Power struct {
	String
}

func NewPower() *Power {
	p := Power{StringWithValidValues("on", "off", "toggle")}
	return &p
}

type Name struct {
	String
}

func NewName() *Name {
	return &Name{}
}

type Version struct {
	String
}

func NewVersion() *Version {
	return &Version{}
}

type Identity struct {
	String
}

func NewIdentity() *Identity {
	return &Identity{}
}

type Model struct {
	String
}

func NewModel() *Model {
	return &Model{}
}

type Manufacturer struct {
	String
}

func NewManufacturer() *Manufacturer {
	return &Manufacturer{}
}
