package wind

import(
	"strconv"
	"math"
)

/*
	风向玫瑰图
*/

// 计算参数
type WdvpArg struct {
	Dtype string
	NAngles int
	IntervalV int
	IntervalP int
}

type WdvpF struct {
	V []float64
	P []float64
	Pf float64
	Vf float64
}
type WdvpFs struct {
	AyD []float64
	AyV []float64
	AyP []float64
	Wdvpfs []WdvpF
	SumV []float64
	SumP []float64
}

func WdvpWindRose(D,V,P []float64, arg WdvpArg) WdvpFs {
	// 参数
	nAngles := arg.NAngles // 划分区间数
	interval := 360.0/float64(nAngles)

	// 设定方向  directions conversion:
	if arg.Dtype=="meteo" {
		D = ArrayMod(ArrayAdd(ArrayTimeN(D, -1.0), 90.0), 360.0) //D = mod(90.0-D, 360.0)
	}

	// angles subdivisons
	D = ArrayMod(ArrayAdd(D, +interval/2.0), 360.0) // D = mod(D, 360.0)
	AyD := ArrayAdd(Linspace(0.0,360.0,nAngles), -0.5*360.0/float64(nAngles))

	stepV := math.Ceil(ArrayMax(V)/float64(arg.IntervalV))
	AyV := []float64{0.5,}
	for i:=1; float64(i)<=stepV; i++ {
		AyV = append(AyV, float64(i*arg.IntervalV))
	}

	maxP := ArrayMax(P)
	catsP := []float64{200.0,250.0,500.0,1000.0,2000.0}
	var reals []float64
	for _, v := range catsP {
		reals = append(reals, math.Abs(maxP/v-7.0))
	}
	index, _ := ArrayMinI(reals)

	arg.IntervalP = int(catsP[index])

	AyP := []float64{}
	for i:=1; float64(i)<=stepV; i++ {
		AyP = append(AyP, float64(i*arg.IntervalP))
	}

	wdsums := map[string]int{}
	wpsums := map[string]float64{}

	wvfs := map[string]map[string]int{}
	wpfs := map[string]map[string]float64{}

	for i, v := range D {
		index := strconv.Itoa(int(v/interval))
		vp := P[i]
		vv := V[i]

		wdsums[index] = wdsums[index] + 1
		wpsums[index] = wpsums[index] + vp

		indexV := strconv.Itoa(getWvIndex(vv, arg.IntervalV))
		indexP := strconv.Itoa(getWpIndex(vp, arg.IntervalP))

		if _, ok := wvfs[index]; !ok {
			wvfs[index] = map[string]int{}
		}
		if _, ok := wpfs[index]; !ok {
			wpfs[index] = map[string]float64{}
		}

		wvfs[index][indexV] = wvfs[index][indexV] + 1
		wpfs[index][indexP] = wpfs[index][indexP] + vp
	}

	allNum := len(D)
	allWp := 0.0
	for _, v := range wpsums {
		allWp = allWp + v
	}

	var wdvpfs []WdvpF
	for i:=0; i<len(AyD); i++ {
		kD := strconv.Itoa(i)
		var wdvpf WdvpF
		wdvpf.Vf = 100*float64(wdsums[kD])/float64(allNum)
		wdvpf.Pf = 100*wpsums[kD]/allWp

		var v, p []float64
		for j:=0; j<len(AyV); j++ {
			kV:= strconv.Itoa(j)
			v = append(v, 100*float64(wvfs[kD][kV])/float64(allNum))
		}
		for j:=0; j<len(AyP); j++ {
			kP:= strconv.Itoa(j)
			p = append(p, 100*wpfs[kD][kP]/allWp)
		}
		wdvpf.V = v
		wdvpf.P = p
		wdvpfs = append(wdvpfs, wdvpf)
	}

	var sumv, sump []float64
	for i:=0; i<len(AyV); i++ {
		var sum float64
		for j:=0; j<len(AyD); j++ {
			sum = sum + wdvpfs[j].V[i]
		}
		sumv = append(sumv, sum)
	}
	for i:=0; i<len(AyP); i++ {
		var sum float64
		for j:=0; j<len(AyD); j++ {
			sum = sum + wdvpfs[j].P[i]
		}
		sump = append(sump, sum)
	}

	wdvpFs := WdvpFs{
		AyD: AyD,
		AyV: AyV,
		AyP: AyP,
		Wdvpfs: wdvpfs,
		SumV: sumv,
		SumP: sump,
	}

	return wdvpFs
}

func getWpIndex(v float64, interval int) int {
	return int(v/float64(interval))
}

