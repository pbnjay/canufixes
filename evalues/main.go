// Command evalues creates a merged .ovlStore/evalues files without any large allocations.
//
// USAGE:
//     cd unitigging/3-overlapErrorAdjustment
//     evalues -o ../oat5.ovlStore/evalues *.oea
//
package main

//
// this is a workaround for the following crash:  (101M overlaps)
//
// -- Starting command on Thu Aug  3 08:31:23 2017 with 142337.4 GB free disk space
//
//     cd unitigging/3-overlapErrorAdjustment
//     /projects/oat_genome/sw/packages/canu/v1.5-73-g80ce789/bin/ovStoreBuild \
//       -G ../oat5.gkpStore \
//       -O ../oat5.ovlStore \
//       -evalues \
//       -L ./oea.files \
//     > ./oea.apply.err 2>&1
// sh: line 5: 130987 Aborted                 (core dumped) /projects/oat_genome/sw/packages/canu/v1.5-73-g80ce789/bin/ovStoreBuild -G ../oat5.gkpStore -O ../oat5.ovlStore -evalues -L ./oea.files > ./oea.apply.err 2>&1
//
// -- Finished on Thu Aug  3 08:31:23 2017 (lickety-split) with 142337.4 GB free disk space
// ----------------------------------------
// ERROR:
// ERROR:  Failed with exit code 134.  (rc=34304)
// ERROR:
//
/////// oea.apply.err:
//
// terminate called after throwing an instance of 'std::bad_alloc'
//   what():  std::bad_alloc
//
// Failed with 'Aborted'; backtrace (libbacktrace):
// AS_UTL/AS_UTL_stackTrace.C::102 in _Z17AS_UTL_catchCrashiP9siginfo_tPv()
// (null)::0 in (null)()
// (null)::0 in (null)()
// (null)::0 in (null)()
// (null)::0 in (null)()
// (null)::0 in (null)()
// (null)::0 in (null)()
// (null)::0 in (null)()
// (null)::0 in (null)()
// (null)::0 in (null)()
// stores/ovStore.C::546 in _ZN7ovStore10addEvaluesERSt6vectorIPcSaIS1_EE()
// stores/ovStoreBuild.C::71 in addEvalues()
// stores/ovStoreBuild.C::468 in main()
// (null)::0 in (null)()
// (null)::0 in (null)()
// (null)::0 in (null)()
//

import (
	"encoding/binary"
	"flag"
	"io"
	"log"
	"os"
	"sort"
)

type OEAFile struct {
	File   *os.File
	LowID  uint32
	HighID uint32
	Length uint64
}

func main() {
	outname := flag.String("o", "merged.oea", "output filename")
	flag.Parse()

	var totalOverlaps uint64
	var infos []*OEAFile

	for _, fn := range flag.Args() {
		f, err := os.Open(fn)
		if err != nil {
			panic(err)
		}
		of := &OEAFile{File: f}

		err = binary.Read(f, binary.LittleEndian, &of.LowID)
		if err != nil {
			panic(err)
		}
		err = binary.Read(f, binary.LittleEndian, &of.HighID)
		if err != nil {
			panic(err)
		}
		err = binary.Read(f, binary.LittleEndian, &of.Length)
		if err != nil {
			panic(err)
		}

		totalOverlaps += of.Length
		infos = append(infos, of)
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].LowID < infos[j].LowID
	})

	if len(infos) == 0 || totalOverlaps == 0 {
		os.Exit(1)
	}
	log.Println("Merging", totalOverlaps, "overlap evalues from", len(infos), "files")

	fout, err := os.Create(*outname)
	if err != nil {
		panic(err)
	}

	//////////
	toPct := float64(len(infos)) / 100.0
	for i, of := range infos {
		log.Printf("%3.2f%% [%d-%d] +%d", float64(i)/toPct, of.LowID, of.HighID, of.Length)

		n, err := io.Copy(fout, of.File)
		if err != nil {
			panic(err)
		}
		if uint64(n) != of.Length*2 {
			log.Fatalf("invalid length written %d bytes written != %d expected", n, of.Length*2)
		}
		of.File.Close()
	}
	fout.Close()
}
