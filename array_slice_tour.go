package main

import "golang.org/x/tour/pic"

func Pic(dx, dy int) [][]uint8 {
	ret := make([][]uint8, dy)
	for j := 0; j < dy; j++ {
		ret[j]=make([]uint8,dx);
		for i := 0; i < dx; i++ {
			ret[j][i]=uint8(j*i-5*i);
		}
	}
	return ret;
}

func main() {
	pic.Show(Pic)
}
