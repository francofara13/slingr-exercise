// Credits to codycollier, this is a modification of his wer repo: https://github.com/codycollier/wer
// for returning substitutions, inclusions and deletions

// WER provides word error rate and related calculations
//
// The word error rate algorithm is implemented per the
// reference found here:
//   https://martin-thoma.com/word-error-rate-calculation/
//
package wer

// Return the minimum of three ints
func minTrio(a, b, c int) int {
	min := a
	if a > b {
		min = b
	}
	if c < min {
		min = c
	}
	return min
}

// Return word error rate and word accuracy for (reference, candidate)
func WER(reference, candidate []string) (float64, float64, int64, int64, int64) {

	lr := len(reference)
	lc := len(candidate)

	OP_OK := 0
	OP_SUB := 1
	OP_INS := 2
	OP_DEL := 3

	// init 2d slice
	D := make([][]int, lr+1)
	backtrace := make([][]int, lr+1)
	for i := range D {
		D[i] = make([]int, lc+1)
		backtrace[i] = make([]int, lc+1)
	}

	// initialization
	for i := 0; i <= lr; i++ {
		for j := 0; j <= lc; j++ {
			if i == 0 {
				D[0][j] = j
				backtrace[0][j] = OP_DEL
			} else if j == 0 {
				D[i][0] = i
				backtrace[i][0] = OP_INS
			}
		}
	}

	// calculation
	for i := 1; i <= lr; i++ {
		for j := 1; j <= lc; j++ {
			if reference[i-1] == candidate[j-1] {
				D[i][j] = D[i-1][j-1]
				backtrace[i][j] = OP_OK
			} else {
				sub := D[i-1][j-1] + 1
				ins := D[i][j-1] + 1
				del := D[i-1][j] + 1
				D[i][j] = minTrio(sub, ins, del)
				if D[i][j] == sub {
					backtrace[i][j] = OP_SUB
				} else if D[i][j] == ins {
					backtrace[i][j] = OP_INS
				} else {
					backtrace[i][j] = OP_DEL
				}
			}
		}
	}

	i := lr
	j := lc
	numSub := 0
	numDel := 0
	numIns := 0
	numCor := 0

	condition := true
	for condition {
		if backtrace[i][j] == OP_OK {
			numCor += 1
			i-=1
			j-=1
		} else if backtrace[i][j] == OP_SUB {
			numSub +=1
			i-=1
			j-=1
		} else if backtrace[i][j] == OP_INS {
			numIns += 1
			j-=1
		} else if backtrace[i][j] == OP_DEL {
			numDel += 1
			i-=1
		}
		if i == 0 || j == 0 {
			condition = false
		}
	}



	wer := float64(D[lr][lc]) / float64(lr)
	wacc := 1.0 - float64(wer)

	return wer, wacc, int64(numSub), int64(numDel), int64(numIns)
}
