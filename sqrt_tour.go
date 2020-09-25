package main

import (
	"fmt"
)

func Sqrt(x float64) float64 {
	z:=1.0;
	preZ:=0.0;
	for (z-preZ > 0.0000001) || (z-preZ < -0.0000001) {
		preZ=z;
		z-=(z*z-x)/(2*z);
		fmt.Println("round:",preZ,z);
	}
	return z;
}

func main() {
	ret:=Sqrt(2);
	fmt.Println(ret,ret*ret-2)
}