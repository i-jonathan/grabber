package main

import "testing"

func BenchmarkMain(b *testing.B) {
	main()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		main()
	}
}
