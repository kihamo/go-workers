// Code generated by "stringer -type=WorkerStatus -trimprefix=WorkerStatus -output worker_status_string.go"; DO NOT EDIT.

package workers

import "strconv"

const _WorkerStatus_name = "WaitProcessCancel"

var _WorkerStatus_index = [...]uint8{0, 4, 11, 17}

func (i WorkerStatus) String() string {
	if i < 0 || i >= WorkerStatus(len(_WorkerStatus_index)-1) {
		return "WorkerStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _WorkerStatus_name[_WorkerStatus_index[i]:_WorkerStatus_index[i+1]]
}
