package crypto

var x86_features = [][]interface{}{
	{"fpu", 0, 0},
	{"vme", 0, 1},
	{"de", 0, 2},
	{"pse", 0, 3},
	{"tsc", 0, 4},
	{"msr", 0, 5},
	{"pae", 0, 6},
	{"mce", 0, 7},
	{"cx8", 0, 8},
	{"apic", 0, 9},
	{"sep", 0, 11},
	{"mtrr", 0, 12},
	{"pge", 0, 13},
	{"mca", 0, 14},
	{"cmov", 0, 15},
	{"pat", 0, 16},
	{"pse-36", 0, 17},
	{"psn", 0, 18},
	{"clflsh", 0, 19},
	{"ds", 0, 21},
	{"acpi", 0, 22},
	{"mmx", 0, 23},
	{"fxsr", 0, 24},
	{"sse", 0, 25},
	{"sse2", 0, 26},
	{"ss", 0, 27},
	{"htt", 0, 28},
	{"tm", 0, 29},
	{"pbe", 0, 31},
	{"sse3", 1, 0},
	{"pclmuldq", 1, 1},
	{"dtes64", 1, 2},
	{"monitor", 1, 3},
	{"ds-cpl", 1, 4},
	{"vmx", 1, 5},
	{"smx", 1, 6},
	{"est", 1, 7},
	{"tm2", 1, 8},
	{"ssse3", 1, 9},
	{"cnxt-id", 1, 10},
	{"cx16", 1, 13},
	{"xtpr", 1, 14},
	{"pdcm", 1, 15},
	{"dca", 1, 18},
	{"sse4.1", 1, 19},
	{"sse4.2", 1, 20},
	{"x2apic", 1, 21},
	{"movbe", 1, 22},
	{"popcnt", 1, 23},
	{"aes", 1, 25},
	{"xsave", 1, 26},
	{"osxsave", 1, 27},
	{"avx", 1, 28},
	{"f16c", 1, 29},
	{"rdrnd", 1, 30},
	{"hypervisor", 1, 31},
	{"fsgsbase", 2, 0},
	{"bmi1", 2, 3},
	{"hle", 2, 4},
	{"avx2", 2, 5},
	{"smep", 2, 7},
	{"bmi2", 2, 8},
	{"erms", 2, 9},
	{"invpcid", 2, 10},
	{"rtm", 2, 11},
	{"mpx", 2, 14},
	{"avx512f", 2, 16},
	{"avx512f", 2, 16},
	{"avx512dq", 2, 17},
	{"rdseed", 2, 18},
	{"adx", 2, 19},
	{"smap", 2, 20},
	{"avx512ifma", 2, 21},
	{"pcommit", 2, 22},
	{"clflushopt", 2, 23},
	{"clwb", 2, 24},
	{"avx512pf", 2, 26},
	{"avx512er", 2, 27},
	{"avx512cd", 2, 28},
	{"sha", 2, 29},
	{"avx512bw", 2, 30},
	{"avx512vl", 2, 31},
}

var armv8_features = [][]interface{}{
	{"neon", 0, 0},
	{"aes", 0, 2},
	{"sha1", 0, 3},
	{"sha2", 0, 4},
	{"pmull", 0, 5},
	{"crc32", 0, 6},
}

// Ref: https://tools.ietf.org/html/draft-agl-tls-chacha20poly1305-04#page-11
// And https://tools.ietf.org/html/draft-strombergson-chacha-test-vectors-00#page-4
var chacha_test_vectors = [][]string{
	{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"0000000000000000",
		"76b8e0ada0f13d90405d6ae55386bd28bdd219b8a08ded1aa836efcc8b770dc7da41597c5157488d7724e03fb8d84a376a43b8f41518a11cc387b669b2ee6586",
	},
	{
		"0000000000000000000000000000000000000000000000000000000000000001",
		"0000000000000000",
		"4540f05a9f1fb296d7736e7b208e3c96eb4fe1834688d2604f450952ed432d41bbe2a0b6ea7566d2a5d1e7e20d42af2c53d792b1c43fea817e9ad275ae546963",
	},
	{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"0000000000000001",
		"de9cba7bf3d69ef5e786dc63973f653a0b49e015adbff7134fcb7df137821031e85a050278a7084527214f73efc7fa5b5277062eb7a0433e445f41e31afab757",
	},
	{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"0100000000000000",
		"ef3fdfd6c61578fbf5cf35bd3dd33b8009631634d21e42ac33960bd138e50d32111e4caf237ee53ca8ad6426194a88545ddc497a0b466e7d6bbdb0041b2f586b",
	},
	{
		"000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f",
		"0001020304050607",
		"f798a189f195e66982105ffb640bb7757f579da31602fc93ec01ac56f85ac3c134a4547b733b46413042c9440049176905d3be59ea1c53f15916155c2be8241a",
	},
}
