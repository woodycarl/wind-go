package wind

type Config struct {
	AllowMinMax bool    // true | false
	RlostMethod string  // avg | random
	CalMethod   string  // notfuse | fuse
	CalHeight   float64 //
	AirDensity  float64
	AutoRevise  bool
	Separate    bool
}

type Station struct {
	System   string
	Version  string
	SensorsR []Sensor
	Logger   Logger
	Site     Site

	Sensors map[string]([]Sensor)

	DataWv   []float64
	DataWd   []float64
	DataWp   []float64
	DataTime []float64

	Am         []Am
	Cm         []Am
	AirDensity float64

	Turbs []float64
	Wss   Wss

	TurbineHeight float64
}

type Result struct {
	ID         string
	S          Station
	D1, D2, RD []Data
}

type Sensor struct {
	Channel      string
	Type         string
	Description  string
	Details      string
	SerialNumber string
	Height       float64
	ScaleFactor  string
	Offset       string
	Units        string

	Rations []Ration

	NotInstalled bool

	Value string
}

type Logger struct {
	Model       string
	Serial      string
	HardwareRev string
}

type Site struct {
	Site          string
	SiteDesc      string
	ProjectCode   string
	ProjectDesc   string
	SiteLocation  string
	SiteElevation float64
	Latitude      string
	Longitude     string
	TimeOffset    string
}

type Ration struct {
	Index     int
	ID        string
	Channel   string
	Rsq       float64
	Slope     float64
	Intercept float64
}

type Am struct {
	My    float64
	Year  float64
	Month float64

	Rwv  int
	Rwd  int
	Twv  int
	Tt   int
	Tp   int
	Cwv  int
	Cwd  int
	Lost int
	All  int
	Sbm  int
	Sr   float64

	NotExist bool
}

type Err struct {
	Err []bool
	Num int
}

type Ws struct {
	XI  int
	YI  int
	XCh string
	YCh string
	XH  float64
	YH  float64
	Ws  float64
}
type Wss struct {
	Ws     []Ws
	A      float64
	B      float64
	R      float64
	Height []float64
	Wv     []float64
}
