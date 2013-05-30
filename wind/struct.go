package wind

type Config struct {
	AllowMinMax bool    // true | false
	RlostMethod string  // avg | random
	CalMethod   string  // notfuse | fuse
	CalHeight   float64 //
	AirDensity  float64
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

	Wvp   []map[string]Mwvp
	Turbs []float64
	Wss   Wss
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
	Height       int
	ScaleFactor  string
	Offset       string
	Units        string

	Rations []Ration
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

	NotExist bool

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
}

type Err struct {
	Err []bool
	Num int
}

type Mwvp struct {
	Wv  float64
	Wp  float64
	Hwv map[string]float64
	Hwp map[string]float64
}

type Ws struct {
	XI  int
	YI  int
	XCh string
	YCh string
	XH  int
	YH  int
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
