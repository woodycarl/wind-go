#SDR

## sensor 

传感器容量 15 | 12 

有时会在前面省略未装的传感器
Channel 1
Channel 2
Channel 7
Channel 8
但在数据部分仍然占位

-----Sensor Information-----

Channel #	1
Type	1
Description	NRG #40 Anem, m/s
Details	090 deg T
Serial Number	SN:
Height	28.4 m
Scale Factor	0.765
Offset	0.35
Units	m/s

Channel #	4
Type	0
Description	No Sensor
Details
Serial Number	SN:
Height	0
Scale Factor	0
Offset	0
Units

Channel #	5
Type	0
Description	No Sensor
Details
Serial Number	--------
Height	------
Scale Factor	0
Offset	0
Units

Channel #	7
Type	3
Description	#200P Wind Vane
Details	000 deg T, logger offset 295
Serial Number	SN:
Height	28.6 m
Scale Factor	0.351
Offset	0
Units	deg

Channel #	9
Type	4
Description	NRG #110S Temp  C
Details	north
Serial Number	SN:
Height	2.5 m
Scale Factor	0.138
Offset	-86.381
Units	C

### height

	Height	28.4 m
	Height	m
	Height	187ft // en
	Height	0
	Height	80   m
	Height	

	1英尺(ft)= 12英寸(in)
	1英寸(in)= 2.54厘米(cm)

### type

	Type	1 风速
	Type	3 风向
	Type	4 温度

### Serial Number

	Serial Number	None
	Serial Number	SN:3012
	Serial Number	00087805

### wv 风速

	Units	m/s
	Units	mph

	1 mph = 1.609344 km/h 
	1 km/h = 0.6213712 mph 

	KM大家都知道是公里的意思，而MPH(miles per hour)大家都俗称“迈”，在国外是“英里”的意思. 

### wd 风向

	Units	deg
	Units	Degrees

### t 温度

	Units	C
	Units	Degrees F
	Units	F

### p 压力

	Units	kPa
	Units	mb
	
	mb = hPa = 0.1kPa

	Type	4
	Description	BP-20 Barom.kPa
	Details	
	Serial Number	SN:5600
	Height	8    m
	Scale Factor	.04255
	Offset	64.832
	Units	kPa

	Type	4
	Description	BP-20 Barom. mb
	Details	
	Serial Number	SN:
	Height	m
	Scale Factor	.426
	Offset	650.031
	Units	mb

### vol 

	Units	volts
	Units	v

	Type	4
	Description	volt   0
	Details	
	Serial Number	SN:
	Height	m
	Scale Factor	.021
	Offset	0
	Units	v

### h 湿度

	Units	%RH

	Type	4
	Description	RH-5 Humidity %RH
	Details	
	Serial Number	SN:
	Height	8    m
	Scale Factor	.098
	Offset	0
	Units	%RH

### 未安装

	Type	0
	Description	No SCM Installed
	Details	
	Serial Number	--------
	Height	------
	Scale Factor	0
	Offset	0
	Units	-----

	Type	4
	Description	Custom
	Details	
	Serial Number	SN:
	Height	m
	Scale Factor	1
	Offset	0
	Units	unit

	Type	0
	Description	No Sensor
	Details	
	Serial Number	
	Height	
	Scale Factor	0
	Offset	0
	Units	

	Type	0
	Description	No Sensor
	Details	
	Serial Number	SN:
	Height	0
	Scale Factor	0
	Offset	0
	Units	

## Site

-----Site Information-----
Site #	0004
Site Desc	LOMA DEL HUESO
Project Code	CNE
Project Desc	OPERACION CNE
Site Location	III REGION
Site Elevation	187
Latitude	S 028? 54.473'
Longitude	W 071? 27.024'
Time offset (hrs)	-4

-----Site Information-----
Site #	1301
Site Desc	Mapes/Airport South
Project Code	New
Project Desc	The site is approximately 3.5 miles north-northwest of the town of Walsenburg
Site Location	Walsenburg, CO
Site Elevation	6105
Latitude	374027N
Longitude	1044804W
Time offset (hrs)	0

	Site Elevation	3136 m
	Site Elevation	6105

	Latitude	374027N
	Latitude	S 028? 54.473'
	Latitude	N 038° 11.591'
	Latitude	0S
	Latitude	N 029 36.282'
	Latitude	N 000 00.000'

	Longitude	W 000 00.000'
	Longitude	E 102 36.835'
	Longitude	0W
	Longitude	1044804W
	Longitude	W 071? 27.024'
	Longitude	W 086° 15.538'

## Logger

-----Logger Information-----
Model #	2333
Serial #	00837
Hardware Rev.	009-000-000


## Date & Time Stamp

	07/16/2011 13:00:00 // en
	01-01-2012 00:00:00 // en
	^(\d{1,2})\/(\d{1,2})\/(\d{4})\s(\d{1,2}):(\d{1,2})(:\d{1,2}|)$

	2011/11/6 Sunday 00:00:00
	2009/05/11 01:00:00
	^(\d{4})\/(\d{1,2})\/(\d{1,2})(\s\w+|)\s(\d{1,2}):(\d{1,2})(:\d{1,2}|)$
	