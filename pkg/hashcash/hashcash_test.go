package hashcash

//func BenchmarkFindNonce(b *testing.B) {
//	ctx := context.Background()
//	hasher := sha1.New()
//	for i := 0; i < b.N; i++ {
//		_, _ = FindNonce(ctx, hasher, []byte("Hello"), 20)
//	}
//}
//
//func BenchmarkFindNonce2(b *testing.B) {
//	ctx := context.Background()
//	hasher := sha1.New()
//	for i := 0; i < b.N; i++ {
//		_, _ = FindNonce2(ctx, hasher, []byte("Hello"), 20)
//	}
//}
