package cachecash

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func decodeDigests(t *testing.T, digests []string) [][]byte {
	var result [][]byte
	for _, s := range digests {
		b, err := hex.DecodeString(s)
		if err != nil {
			t.Fatalf("failed to decode expected block digest: %v", err)
		}
		result = append(result, b)
	}
	return result
}

func assertBlockDigests(t *testing.T, objs map[string]ContentObject, path string, expected [][]byte) {
	obj, ok := objs[path]
	if !assert.True(t, ok, "no object with expected path") {
		return
	}

	if !assert.Equal(t, len(expected), obj.BlockCount(), "object contains unexpected number of blocks") {
		return
	}

	for i := 0; i < obj.BlockCount(); i++ {
		d, err := obj.BlockDigest(uint32(i))
		assert.Nil(t, err, "failed to compute digest")

		// N.B.: The hex encoding is only to make output more pleasant.
		assert.Equal(t, hex.EncodeToString(expected[i]), hex.EncodeToString(d), fmt.Sprintf("digest mismatch for block %d", i))
	}
}

func TestContentCatalog(t *testing.T) {

	blockDigestsFile0 := decodeDigests(t, []string{
		"cb2070d0b083e73d4e445a49dfd2fffd780168e28cd9263a532292183f38b208a280a415e566bba1d36c85282206e8d2",
		"d1418d4cc3f34a43704553ff768715be86eb396accd839dfc42bfd3317122fd60d5f6944eba3e9b5ef278e28b0cb0d84",
		"d3f908125c90453f056cb46ab32a09dfceceaab41f53053ccd586ed786ec476de9608193db798a04f9bffa1528184b27",
		"29edb8e612431aaf027b7b00b9e2dfd158f6797f9f46af2447ae4362ebc0136bfc27b584b6f5dcd4fb958c67cb275edb",
		"a3679d76fd2715ce15b07c31f6b6757af8acd93c70361c4c07b4f3c00d4f74e8c84ad26001278daca7ba553c7cf1087b",
		"386a810905375cff2778f30a0ed849df0de318d39b6968089351a1e238b1cb4603b1da22681ba530029d305bf5a4e83c",
		"bf646957bd859670c54aa55dbb7f94039f1bea52574b4f80936b34fe6e834c45b99063ccb3873c2d81eb3974e72669c5",
		"ecd9e929ae93766d1805fde1d01bc7c5ebcf9d48e3a36b9360688e94572a92a9a0d2890b7a29d072d82512daf2cc82fb",
		"1db162c3819038a7a48c8007ad34e34bb284596912a3465efbf29a34a5bf544516a30a5dcaf1c342336ee9b2ce9990aa",
		"a3fda9af1fe3a36bb27893259e2e5a0f186c3f1c817129104cdbdb64cb7ccf690732ebbfa009b0abcd5d31a7af6df7f3",
		"2b21b71a827c192cb94384390c2c19eea6dd6031d63ef102fa401315d0f519da871f4c9ce3a7d5979dbb88650c62a441",
		"f9c51f4c6916d2462e5d5f439a453e539c7cae08875814b4989da8840ab59b41dbf71a72f15edda1d7d8a3bb12f56e1b",
		"05403611653c48b7c5faa9069dbd9f0c3aba62dc5f06869a0d9877a07d7002157465c6c2f19303febbed4dbb1294644b",
		"df7f28c910af24fb3e04a58005d5a88ea9c343bb1417f68bfee22d736b9d5ae0ecf433262ffe0dfc008b8ba34ae699fc",
		"095abf0fb5ccdf755a3e5f02031236c1d79eeafcbcbe3a0cc4e68dabf87f6975d77111828e582974d0c65493af0783d7",
		"6270ac323102e2fd59a39d7a1bf77af000c23f8992d09c90db99102be464daebb2b875b53484c7f7abe361ffaa5a8b95",
		"7e2569cc9c403a1083590697b1fca337ea762852386e560fa7794029eff140e8f9f59637da01fbf18a55c6374adbc15d",
		"f567074431eca2558a9586c53fbe1c90d281f076e1ffca5f61b59e527a2e0117823d552b567bafd1ee93c0605adf7b85",
		"2899d1ae2b882ed6de124d4914adad5e0d48403bc7862ea20c4b31eff4708b7744a8fe43fece1d6a84b150acb4c5fa82",
	})

	blockDigestsFile1 := decodeDigests(t, []string{
		"400be9aa3d312581c291865696e5e4827c8f8c8a6f870253e3964e1fd6c0f74ed3afbd348c346b88cf3c5db33bcf6261",
		"8868c6b08153804aa0a8663f8342c9d078b030c610630dd3c7f3d3272c0e78e0f7a502beb5de0c64e1c9e48a1dcb5f44",
		"a66da9ef36b73a5a9d5e07e6becf8df10aded5417db4bd5dc61e3272555975d92144c1c4cbd5427b250928a398704ba4",
		"9543090670bd77e553e03c33a56d231196c9e1894211fac4aed54a630d253255e11a652333a2bad20eff754b34763b5c",
		"05c6fcab3d088b510cf6d1640651a18c99ffa2f03408f2898151a363d67cbb210e3046474c83c1ef479e04e9c8f1a952",
		"a335fbefe0b80966a8120807f958b611b1e33612e0dcdba4e858dce16e88a7f2dd45421ad34ad159cc5dceec12457b03",
		"3295d2abd29a87be694cb3ce5f687a091052cadf6a9681ce93939d91d5936736d9965ff449fe7155a5dc9fac9016f6e9",
		"d5c392dd389d1b1aab95da70fa6d31948016048a72c289defe3c28ba960bd4039c6fb79d8505b7046c497fa4f61050da",
		"62dc3846ca238d9f4f2a886fbb5287e3232e413c663870cea7f22f38ae872874cb06b466a24bfb5e191032fc3ab6711d",
		"80ed7fd21f4bfe83027d80e7b13224b29cd09fa79692844d431cfcf71ed6bf97ab5d9b35e7974e100e19c7d7bdd15f16",
		"0339a693586b15403f53c962c43c8e1e96b62253a3cfbf997631ecfca128fefa66af79c7ae1dae30fbc6a7306cbd6707",
		"c79040cdd867f0f40a10afcb9b19cec0858175efe50e74fb2c4a4d6cf53e93221558e208f4be748c584e8344125ad1ac",
		"4839363cf7d384c359881d35d36b79eaff9680a9140e0a9781b56db4b8f39e31803a14416a828ff63e84c007098d4a21",
		"fc567ed1b0c912d6fa81cf9f0dbe35202922c52339d6ef745e2276ab13860c7238f3598179607c1934a539693989b915",
		"b5cb8717ce5e9eb52d73c2aed5cfc860328df88fb2bdefdc8ac89cf528505ec53a69581bcb8175dc4ab4595e60c3f257",
		"64a4e41bc1f575f4d8dbf99086a0d19f48481eae3d08be964b0d833abe8c853381078916461f0c848c8c7f1b935ed0d4",
		"e335c7130128b5c341e57f6c8b4aa98781abe6c1c3fd87dfe1c8f2493062800b54cfc30a71f8ea52db68ff631c7380c1",
		"a24960ba26031967cc6e86fa39a7c0cb6323aaef3d3ad22a2a0ff17017e4ddf0a9adc18440e5290fb74d2233f3982936",
		"c3912f03a046c58f08e11e84e631f78831d915dd5e2245906bcaa33cde284a576ebb8055dd811675f96437627f43ea19",
	})

	cat, err := NewCatalogFromDir("testdata/content")
	if err != nil {
		t.Fatalf("error creating catalog: %v", err)
	}

	objs := cat.ObjectsByPath()

	assert.Equal(t, 2, len(objs))
	assertBlockDigests(t, objs, "/file0.bin", blockDigestsFile0)
	assertBlockDigests(t, objs, "/file1.bin", blockDigestsFile1)
}
