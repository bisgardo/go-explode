package explode

import "testing"

func Benchmark__increasing_group_sizes_on_same_level(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Explode("a{b}{c,d}{e,f,g}{h,i,j,k}{l,m,n,o,p}{q,r,s,t,v,w}")
	}
}
