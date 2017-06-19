// Command mergeclears merges multiple .clear files produced from CANU's
// splitReads commands (run in parallel). CANU's code doesn't have a way to
// merge these files, and crashes for unknown reasons during splitReads.
//
// We externally parallelized canu's splitReads step to avoid the crash, so
// we needed this code to merge the results back together.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
)

func main() {
	outname := flag.String("o", "merged.clear", "output clear file")
	flag.Parse()

	var numReads uint32
	var alldata []uint32
	for ax, fn := range flag.Args() {
		fmt.Printf("%d/%d %s\n", ax, flag.NArg(), fn)
		f, err := os.Open(fn)
		if err != nil {
			panic(err)
		}
		var nreads uint32
		err = binary.Read(f, binary.LittleEndian, &nreads)
		if err != nil {
			panic(err)
		}
		if numReads == 0 {
			numReads = nreads
		}
		if numReads != nreads {
			panic("numreads doesn't match in all files")
		}
		data := make([]uint32, nreads*2+2)
		err = binary.Read(f, binary.LittleEndian, data)
		if err != nil {
			panic(err)
		}
		f.Close()

		if alldata == nil {
			alldata = data
			continue
		}

		for i := uint32(0); i <= nreads; i++ {
			if alldata[i] == ^uint32(0) {
				// deleted, don't replace it
				continue
			}

			if data[i+nreads+1] == ^uint32(0) || data[i] == ^uint32(0) {
				// now it's deleted
				alldata[i] = ^uint32(0)
				alldata[i+nreads+1] = ^uint32(0)
				continue
			}

			// is current begin pt before new begin point?
			// is current end pt after new end point?
			if alldata[i] < data[i] || alldata[i+nreads+1] > data[i+nreads+1] {
				// replace both
				alldata[i] = data[i]
				alldata[i+nreads+1] = data[i+nreads+1]
			}
		}
	}

	f, err := os.Create(*outname)
	if err != nil {
		panic(err)
	}
	err = binary.Write(f, binary.LittleEndian, numReads)
	if err != nil {
		panic(err)
	}
	err = binary.Write(f, binary.LittleEndian, alldata)
	if err != nil {
		panic(err)
	}
	f.Close()
}
