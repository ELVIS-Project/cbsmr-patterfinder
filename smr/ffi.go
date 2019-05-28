package main
/*

import (
	pb "../proto"
	"fmt"
	"sort"
	"unsafe"
	"encoding/gob"
	"bytes"
	"errors"
)

// #cgo LDFLAGS: -L ./helsinki-ttwi/w2 -l w2
// #include "./helsinki-ttwi/w2/w2.h"
import "C"

const (
	WINDOW = 20
)

type CScore *C.struct_Score

func PrintScore(s *C.struct_Score) {
	println(fmt.Sprintf("Score with %v notes and %v vecs:", s.num_notes, s.num_vectors))
	vecs := (*[1 << 30]C.struct_IntraVector)(unsafe.Pointer(s.vectors))
	for i := 0; (C.int)(i) < s.num_vectors; i++ {
		cv := vecs[i]
		println(fmt.Sprintf("x: %v, y: %v, si: %v, ei; %v", cv.x, cv.y, cv.startIndex, cv.endIndex))
	}
}

type Score struct {
	Vectors []vector
	NumNotes int
}

func (s Score) Encode() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(s)
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func DecodeScore(input []byte) (s Score, err error) {
	buf := bytes.NewBuffer(input)
	decoder := gob.NewDecoder(buf)
	err = decoder.Decode(&s)
	if err != nil {
		return Score{}, err
	}
	return
}

type vector struct {
	X          float64
	Y          int32
	StartIndex uint32
	EndIndex   uint32
}

type byHeightThenIndex []vector

func (vs byHeightThenIndex) Len() int {
	return len(vs)
}
func (vs byHeightThenIndex) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}
func (vs byHeightThenIndex) Less(i, j int) bool {
	if vs[i].Y < vs[j].Y {
		return true
	} else if vs[i].Y == vs[j].Y {
		return vs[i].StartIndex <= vs[j].StartIndex
	} else {
		return false
	}
}


func VecsFromNotes(Notes *pb.Notes) (vecs []vector) {
	notes := Notes.Notes

	for i, _ := range notes {
		for j := i + 1; j < min(i+WINDOW, len(notes)); j++ {
			cvec := vector{
				(float64)(notes[j].Onset - notes[i].Onset),
				notes[j].PitchB40 - notes[i].PitchB40,
				notes[i].PieceIdx,
				notes[j].PieceIdx,
			}
			vecs = append(vecs, cvec)
		}
	}
	sort.Sort(byHeightThenIndex(vecs))
	return
}

func InitScoreFromIndexerResp(resp *pb.Notes) (s CScore) {
	vecs := VecsFromNotes(resp)
	s = InitScoreFromVectors(len(resp.Notes), vecs)
	return
}

func InitScoreFromVectors(numNotes int, vecs []vector) (s CScore) {
	CVectors := (*C.struct_IntraVector)(C.malloc(C.sizeof_struct_IntraVector * (C.ulong)(len(vecs))))
	GoCVectors := (*[1 << 30]C.struct_IntraVector)(unsafe.Pointer(CVectors))
	for i, v := range vecs {
		CVec := (C.struct_IntraVector)(C.NewIntraVector(
			(C.float)(v.X),
			(C.int)(v.Y),
			(C.int)(v.StartIndex),
			(C.int)(v.EndIndex),
		))

		GoCVectors[i] = CVec
	}

	s = C.initScoreFromVectors(
		(C.int)(numNotes), (C.int)(len(vecs)),
		CVectors,
	)
	return
}
*/

func min(a int, b int) (minimum int) {
	if a < b {
		return a
	}
	return b
}

func search(pattern CScore, target CScore) (arrays [][]uint32, err error) {

	if pattern.num_notes < 2 {
		return arrays, errors.New("pattern must be at least two notes long")
	}

	result := CSearch(pattern, target)

	return resultToIntArrays(result, pattern), nil
}

func InitScoreFromCsv(vector_csv string) (score *C.struct_Score) {
	score = C.init_score(C.CString(vector_csv))
	return score
}

func CSearch(pattern *C.struct_Score, target *C.struct_Score) (result *C.struct_Result) {
	result = (*C.struct_Result)(C.malloc(C.sizeof_struct_Result))
	C.search(
		(*C.struct_Score)(unsafe.Pointer(pattern)),
		(*C.struct_Score)(unsafe.Pointer(target)),
		(*C.struct_Result)(unsafe.Pointer(result)),
	)
	return
}

func resultToIntArrays(result *C.struct_Result, pattern *C.struct_Score) (arrays [][]uint32) {

	chains := (*[1 << 30]*C.int)(unsafe.Pointer(result.chains))
	for i := 0; (C.int)(i) < result.num_occs; i++ {
		// weird.. can't cast a pointer to a larger array with variable size at compile time?
		chain := (*[1 << 10]C.int)(unsafe.Pointer(chains[i]))
		arr := make([]uint32, pattern.num_notes)

		for j := 0; (C.int)(j) < pattern.num_notes; j++ {
			arr[j] = (uint32)(chain[j])
			if chain[j] == 0 {
				break
			}
		}

		arrays = append(arrays, arr)
	}

	return
}

/*
func InitScoreFromPbVectors
			cvec := vector{
				(float64)(notes[j].Onset - notes[i].Onset),
				notes[j].PitchB40 - notes[i].PitchB40,
				notes[i].PieceIdx,
				notes[j].PieceIdx,
			}

type pbVecHeightIndex []*pb.Vector

func (vs pbVecHeightIndex) Len() int {
	return len(vs)
}
func (vs pbVecHeightIndex) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}
func (vs pbVecHeightIndex) Less(i, j int) bool {
	yi := vs[i].End.PitchB40 - vs[i].Start.PitchB40
	yj := vs[j].End.PitchB40 - vs[j].Start.PitchB40

	if yi < yj {
		return true
	} else if yi == yj {
		return vs[i].Start.PieceIdx <= vs[j].Start.PieceIdx
	} else {
		return false
	}
}

func PbVecsFromPbNotes(Notes *pb.Notes) (vecs []*pb.Vector) {
	notes := Notes.Notes

	for i, _ := range notes {
		for j := i + 1; j < min(i+WINDOW, len(notes)); j++ {
			pbVec := &pb.Vector{
				Start: notes[i],
				End: Note[j],
			}
			vecs = append(vecs, pbVec)
		}
	}
	sort.Sort(pbVecHeightIndex(vecs))
	return
}



func recurseChain(kEntry *C.struct_KEntry) []*C.struct_IntraVector {
	if kEntry.y == nil {
		return [2]*C.struct_IntraVector{kEntry.targetVec.startIndex, kEntry.targetVec.endIndex}
	} else {
		return append(recurseChain(KEntry.y), [1]*C.struct_IntraVector{KEntry.targetVec.endIndex})
	}
}
func kTableToOccurrence(result *C.struct_Result, kTableLength int) pb.Occurrence {
	chains := [][]C.struct_IntraVector{}
	table := ([]*C.struct_KEntry)(result.table)

	for i := 0; i < kTableLength; i++ {
		if result.table[i].w > 0 {
			chains = append(chains, recursively_extract_chain(KTable[i]))
		}
	}

	return chains
}
func twoCombinations(iterable interface{}) (combos [][]interface{}) {
	for i, a := range iterable {
		for j, b := range iterable {
			combos = append(combo, []interface{a, b})
		}
	}

}

func (vec NoteVector) toIntArray() (arr [8]int) {
	arr[0] = (C.float)(vec.b.Onset - vec.a.Onset)
	arr[1] = (C.int)(vec.b.PitchB40 - vec.a.PitchB40)
	arr[2] = (C.int)(vec.a.pieceIdx)
	arr[3] = (C.int)(vec.b.pieceIdx)
	arr[4] = (C.int)(vec.a.PitchB40)
	arr[5] = (C.int)(vec.b.PitchB40)
	arr[6] = (C.int)(arr[0])
	arr[7] = (C.int)(arr[0])
}

func vecsToIntArray(vecs []NoteVector) (arr []int) {
	arr = make([]int, len(vecs)*8)
	for i, vec := range vecs {
		println("%v", i)
		vecArray := vec.toIntArray()
		for i := 0; i < 8; i++ {
			arr[i] = vecArray[i]
		}
		println("%v", arr)
	}
	return
}

func InitScoreWithIntArray(notes []*pb.Note) (C_score *C.struct_Score) {
	vecs := VecsFromNotes(notes)
	vecsAsIntArray := vecsToIntArray(vecs)
	C_score = C.init_score_from_int_array((C.int)(len(notes)), (C.int)(len(vecs)), (*C.int)(unsafe.Pointer(&vecsAsIntArray)))
}

func InitScoreFromNotes(notes []*pb.Note) (C_score *C.struct_Score) {
	var C_vecs []*C.struct_IntraVector
	vecs := VecsFromNotes(notes)

	//mc_vecs := (*C.struct_IntraVector)(C.malloc(C.size_t(len(vecs))))
	//mc_vecs_ptrs := (*[]C.struct_IntraVector)(unsafe.Pointer(mc_vecs))[:len(vecs):len(vecs)]

	for _, vec := range vecs {
		//mc_vec := (C.struct_IntraVector)(C.malloc(C.sizeof_struct_IntraVector))
		/*
			C_vecs = append(C_vecs, &C.struct_IntraVector{
				x:             (C.float)(vec.b.Onset - vec.a.Onset),
				y:             (C.int)(vec.b.PitchB40 - vec.b.PitchB40),
				startIndex:    (C.int)(vec.b.PitchB40 - vec.b.PitchB40),
				endIndex:      (C.int)(vec.b.PitchB40 - vec.b.PitchB40),
				diatonicDiff:  (C.int)(vec.b.PitchB40 - vec.b.PitchB40),
				chromaticDiff: (C.int)(vec.b.PitchB40 - vec.b.PitchB40),
			})
		x := (C.float)(vec.b.Onset - vec.a.Onset)
		y := (C.int)(vec.b.PitchB40 - vec.a.PitchB40)
		sI := (C.int)(vec.a.PieceIdx)
		eI := (C.int)(vec.b.PieceIdx)
		sP := (C.int)(vec.a.PitchB40)
		eP := (C.int)(vec.b.PitchB40)
		dD := (C.int)(y)
		cD := (C.int)(y)
		C_vecs = append(C_vecs, C.NewIntraVector(x, y, sI, eI, sP, eP, dD, cD))
	}

	println("init score")
	C_score = C.init_score_from_vectors(
		(C.int)(len(notes)),
		(C.int)(len(vecs)),
		(*C.struct_IntraVector)(unsafe.Pointer(&C_vecs[0])),
	)
		C_score = (*C.struct_Score)(C.malloc(C.sizeof_struct_Score))
		C_score = &C.struct_Score{
			vectors:     (*C.struct_IntraVector)(unsafe.Pointer(&C_vecs[0])),
			num_notes:   (C.int)(len(notes)),
			num_vectors: (C.int)(len(vecs)),
		}

	return
}


type NoteVector struct {
	a *pb.Note
	b *pb.Note
}

*/
