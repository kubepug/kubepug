BenchmarkOldParser/version_v1.19.5-16         	      26	  43250157 ns/op	19643921 B/op	  226108 allocs/op
BenchmarkOldParser/version_v1.23.4-16         	      33	  36179402 ns/op	17209586 B/op	  196236 allocs/op
BenchmarkOldParser/version_v1.27.2-16         	      30	  37116931 ns/op	18545998 B/op	  195597 allocs/op

Just moved to jjson unmarshaller
BenchmarkNewParser/version_v1.19.5-16         	      34	  33165625 ns/op	 8362156 B/op	   44459 allocs/op
BenchmarkNewParser/version_v1.23.4-16         	      34	  29881963 ns/op	 7570440 B/op	   41523 allocs/op
BenchmarkNewParser/version_v1.27.2-16         	      32	  31324179 ns/op	 7971592 B/op	   39698 allocs/op

Improved struct and loops
BenchmarkNewParser/version_v1.19.5-16         	      37	  30764782 ns/op	 5697117 B/op	    4641 allocs/op
BenchmarkNewParser/version_v1.23.4-16         	      42	  26909969 ns/op	 5047905 B/op	    4104 allocs/op
BenchmarkNewParser/version_v1.27.2-16         	      37	  27897261 ns/op	 5514952 B/op	    4024 allocs/op

Using goccy json
BenchmarkParser/version_v1.19.5-16         	     156	   6719525 ns/op	10979240 B/op	    3863 allocs/op
BenchmarkParser/version_v1.23.4-16         	     207	   5917886 ns/op	 9719129 B/op	    3475 allocs/op
BenchmarkParser/version_v1.27.2-16         	     194	   6370366 ns/op	10671241 B/op	    3489 allocs/op


