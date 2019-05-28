package main

import (
	"context"
	pb "../proto"
	"sort"
)

func max(i, j int) int {
	if i < j {
		return j
	} else {
		return i
	}
}

type rankTrivial []*pb.Occurrence

func (occs rankTrivial) Len() int {
	return len(occs)
}
func (occs rankTrivial) Swap(i, j int) {
	occs[i], occs[j] = occs[j], occs[i]
}
func (occs rankTrivial) Less(i, j int) bool {
	// Most relevant come first
	if len(occs[i].Notes) < len(occs[j].Notes)  {
		return true
	} else if len(occs[i].Notes) > len(occs[j].Notes) {
		return false
	}

	// now len(occs[i].Notes) == len(occs[j].Notes)
	var sum_i uint32
	var sum_j uint32
	for k := 1; k < len(occs[i].Notes); k++ {
		sum_i += occs[i].Notes[k] - occs[i].Notes[k-1]
		sum_j += occs[j].Notes[k] - occs[j].Notes[k-1]
	}

	// Prefer more compact occurrences
	if sum_i < sum_j {
		return true
	} else if sum_i > sum_j {
		return false
	} else {
		// If all else equal, sort on piece id for deterministic results
		return occs[i].Pid < occs[j].Pid
	}
}


func Search(pieceMap map[uint32]CScore, req *pb.Notes) (occs *pb.Occurrences, err error) {
	occs = &pb.Occurrences{}

	vecs := VecsFromNotes(req)

	queryScore := InitScoreFromVectors(len(req.Notes), vecs)

	for pieceID, cscore := range pieceMap {
		intArrays, err := search(queryScore, cscore)
		if err != nil {
			return &pb.Occurrences{}, err
		}
		for _, arr := range intArrays {
			occs.Occurrences = append(occs.Occurrences, &pb.Occurrence{Pid: pieceID, Notes: arr})
		}
	}

	if len(occs.Occurrences) == 0 {
		return &pb.Occurrences{}, nil
	}

	sortedOccs := rankTrivial(occs.Occurrences)
	sort.Sort(sortedOccs)
	occs.Occurrences = sortedOccs

	return occs, nil
}


func (s SmrServer) Search(ctx context.Context, req *pb.Notes) (occs *pb.Occurrences, err error) {
	occs, err = Search(s.pieceMap, req)
	if err != nil {
		return &pb.Occurrences{}, err
	}
	return occs, nil
}
