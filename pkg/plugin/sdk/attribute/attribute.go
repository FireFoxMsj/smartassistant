package attribute

type Power struct {
	String
}

func NewPower() *Power {
	p := Power{StringWithValidValues("on", "off", "switch")}
	return &p
}

type ColorTemp struct {
	Int
}
type Brightness struct {
	Int
}
type Name struct {
	String
}

type Version struct {
	String
}

type Identity struct {
	String
}

type Model struct {
	String
}
type Manufacturer struct {
	String
}
